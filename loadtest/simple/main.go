package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IsaacDSC/featureflag/sdk/featureflag"
)

type FeatureFlagPayload struct {
	FlagName    string `json:"flag_name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type Stats struct {
	totalOperations   int64
	successOperations int64
	errorOperations   int64
	totalLatency      int64
	minLatency        int64
	maxLatency        int64
}

func main() {
	serverURL := "http://localhost:3000"
	workers := 50
	duration := 30 * time.Second
	opsPerSecond := 100

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         TESTE DE CARGA - FEATURE FLAG SERVICE              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. Verificar se servidor está disponível
	fmt.Println("🔍 Verificando conectividade com o servidor...")
	resp, err := http.Get(serverURL + "/ping")
	if err != nil {
		fmt.Printf("❌ Erro ao conectar no servidor: %v\n", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Servidor retornou status: %d\n", resp.StatusCode)
		return
	}
	fmt.Println("✅ Servidor está online!")
	fmt.Println()

	// 2. Criar feature flags necessárias
	fmt.Println("📝 Criando feature flags para teste...")

	flags := []FeatureFlagPayload{
		{FlagName: "new_name", Description: "new_description", Active: true},
		{FlagName: "new_name1", Description: "new_description", Active: true},
	}

	for _, flag := range flags {
		if err := createFeatureFlag(serverURL, flag); err != nil {
			fmt.Printf("⚠️  Erro ao criar flag '%s': %v\n", flag.FlagName, err)
		} else {
			fmt.Printf("✅ Flag '%s' criada com sucesso\n", flag.FlagName)
		}
	}
	fmt.Println()

	// 3. Inicializar SDK e Listener
	fmt.Println("🚀 Inicializando SDK e Listener...")
	ctx := context.Background()
	ff := featureflag.NewFeatureFlagSDK(serverURL)

	go func() {
		_, err := ff.Listenner(ctx)
		if err != nil {
			panic(err)
		}
	}()

	// Aguardar listener inicializar
	time.Sleep(2 * time.Second)
	fmt.Println("✅ SDK inicializado")
	fmt.Println()

	// 4. Executar teste de carga
	fmt.Println("⚡ Iniciando teste de carga...")
	fmt.Printf("   Workers: %d\n", workers)
	fmt.Printf("   Duração: %s\n", duration)
	fmt.Printf("   Taxa alvo: ~%d ops/s\n", opsPerSecond*4) // 4 operações por ciclo
	fmt.Println()

	stats := &Stats{
		minLatency: int64(^uint64(0) >> 1), // Max int64
	}

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	// Iniciar workers
	opsPerWorker := opsPerSecond / workers
	if opsPerWorker < 1 {
		opsPerWorker = 1
	}

	startTime := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(i+1, ff, stats, &wg, stopChan, opsPerWorker)
	}

	// Goroutine para mostrar progresso
	go showProgress(stats, startTime)

	// Aguardar duração do teste
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()

	elapsed := time.Since(startTime)

	// 5. Exibir resultados
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    RESULTADOS DO TESTE                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	printFinalReport(stats, elapsed)
}

func worker(id int, ff *featureflag.FeatureFlagSDK, stats *Stats, wg *sync.WaitGroup, stopChan chan bool, opsPerSecond int) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(opsPerSecond))
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			performOperations(ff, stats)
		}
	}
}

func performOperations(ff *featureflag.FeatureFlagSDK, stats *Stats) {
	// Operação 1: GetFeatureFlag("invalid_ff").WithDefault(true)
	start := time.Now()
	isActive := ff.GetFeatureFlag("invalid_ff").WithDefault(true)
	recordOperation(stats, start, isActive, nil)

	// Operação 3: GetFeatureFlag("new_name").Err()
	start = time.Now()
	isActive2, err := ff.GetFeatureFlag("new_name").Err()
	recordOperation(stats, start, isActive2, err)

	// Operação 4: GetFeatureFlag("new_name1").Val()
	start = time.Now()
	isActive3 := ff.GetFeatureFlag("new_name1").Val()
	recordOperation(stats, start, isActive3, nil)
}

func recordOperation(stats *Stats, start time.Time, result bool, err error) {
	latency := time.Since(start).Microseconds()

	atomic.AddInt64(&stats.totalOperations, 1)
	atomic.AddInt64(&stats.totalLatency, latency)

	if err != nil {
		atomic.AddInt64(&stats.errorOperations, 1)
	} else {
		atomic.AddInt64(&stats.successOperations, 1)
	}

	// Atualizar min latency
	for {
		min := atomic.LoadInt64(&stats.minLatency)
		if latency >= min || atomic.CompareAndSwapInt64(&stats.minLatency, min, latency) {
			break
		}
	}

	// Atualizar max latency
	for {
		max := atomic.LoadInt64(&stats.maxLatency)
		if latency <= max || atomic.CompareAndSwapInt64(&stats.maxLatency, max, latency) {
			break
		}
	}
}

func showProgress(stats *Stats, startTime time.Time) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		elapsed := time.Since(startTime)
		total := atomic.LoadInt64(&stats.totalOperations)
		success := atomic.LoadInt64(&stats.successOperations)
		errors := atomic.LoadInt64(&stats.errorOperations)

		rate := float64(total) / elapsed.Seconds()
		successRate := float64(0)
		if total > 0 {
			successRate = float64(success) / float64(total) * 100
		}

		avgLatency := int64(0)
		if total > 0 {
			avgLatency = atomic.LoadInt64(&stats.totalLatency) / total
		}

		fmt.Printf("⚡ [%5.1fs] Ops: %7d | Rate: %7.1f/s | Success: %6.2f%% | Avg: %6.2fms | Erros: %d\n",
			elapsed.Seconds(),
			total,
			rate,
			successRate,
			float64(avgLatency)/1000.0,
			errors,
		)
	}
}

func printFinalReport(stats *Stats, elapsed time.Duration) {
	total := atomic.LoadInt64(&stats.totalOperations)
	success := atomic.LoadInt64(&stats.successOperations)
	errors := atomic.LoadInt64(&stats.errorOperations)
	totalLatency := atomic.LoadInt64(&stats.totalLatency)
	minLat := atomic.LoadInt64(&stats.minLatency)
	maxLat := atomic.LoadInt64(&stats.maxLatency)

	avgLatency := int64(0)
	if total > 0 {
		avgLatency = totalLatency / total
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	rate := float64(total) / elapsed.Seconds()

	fmt.Printf("📊 Estatísticas Gerais:\n")
	fmt.Printf("   Duração total: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("   Operações totais: %d\n", total)
	fmt.Printf("   Operações/segundo: %.2f\n", rate)
	fmt.Printf("   Ciclos completos: %d (4 ops/ciclo)\n", total/4)
	fmt.Println()

	fmt.Printf("⏱️  Latências:\n")
	fmt.Printf("   Mínima: %.3fms\n", float64(minLat)/1000.0)
	fmt.Printf("   Média: %.3fms\n", float64(avgLatency)/1000.0)
	fmt.Printf("   Máxima: %.3fms\n", float64(maxLat)/1000.0)
	fmt.Println()

	fmt.Printf("✅ Resultados:\n")
	fmt.Printf("   Sucesso: %d (%.2f%%)\n", success, successRate)
	fmt.Printf("   Erros: %d (%.2f%%)\n", errors, 100-successRate)
	fmt.Println()

	if successRate >= 99.9 {
		fmt.Println("🎉 EXCELENTE! Taxa de sucesso >= 99.9%")
	} else if successRate >= 95.0 {
		fmt.Println("✅ BOM! Taxa de sucesso >= 95%")
	} else {
		fmt.Println("⚠️  ATENÇÃO! Taxa de sucesso abaixo de 95%")
	}

	fmt.Println()
	fmt.Println("🎯 Teste de carga concluído!")
}

func createFeatureFlag(serverURL string, payload FeatureFlagPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", serverURL+"/featureflag", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return nil
}
