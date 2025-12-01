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
	toggleCount       int64
	validationErrors  int64
	correctValues     int64
}

type SharedState struct {
	mu              sync.RWMutex
	currentNewName  bool
	currentNewName1 bool
	lastToggleTime  time.Time
}

func main() {
	serverURL := "http://localhost:3000"
	workers := 50
	duration := 30 * time.Second
	opsPerSecond := 100

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║      TESTE DE CARGA 2 - FEATURE FLAG WITH TOGGLE          ║")
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

	// 4. Executar teste de carga com toggle
	fmt.Println("⚡ Iniciando teste de carga com alternância de flags...")
	fmt.Printf("   Workers: %d\n", workers)
	fmt.Printf("   Duração: %s\n", duration)
	fmt.Printf("   Taxa alvo: ~%d ops/s\n", opsPerSecond*3) // 3 operações por ciclo
	fmt.Println()

	stats := &Stats{
		minLatency: int64(^uint64(0) >> 1), // Max int64
	}

	sharedState := &SharedState{
		currentNewName:  true,
		currentNewName1: true,
		lastToggleTime:  time.Now(),
	}

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	// Goroutine para alternar o valor da flag
	go toggleWorker(serverURL, stopChan, stats, sharedState)

	// Iniciar workers
	opsPerWorker := opsPerSecond / workers
	if opsPerWorker < 1 {
		opsPerWorker = 1
	}

	startTime := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(i+1, ff, stats, &wg, stopChan, opsPerWorker, sharedState)
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

func toggleWorker(serverURL string, stopChan chan bool, stats *Stats, sharedState *SharedState) {
	ticker := time.NewTicker(2 * time.Second) // Alterna a cada 2 segundos
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			// Alternar o valor
			sharedState.mu.Lock()
			sharedState.currentNewName = !sharedState.currentNewName
			newValue := sharedState.currentNewName
			sharedState.lastToggleTime = time.Now()
			sharedState.mu.Unlock()

			// Atualizar new_name
			payload := FeatureFlagPayload{
				FlagName:    "new_name",
				Description: "new_description",
				Active:      newValue,
			}

			if err := createFeatureFlag(serverURL, payload); err != nil {
				fmt.Printf("⚠️  Erro ao alternar new_name para %v: %v\n", newValue, err)
			} else {
				atomic.AddInt64(&stats.toggleCount, 1)
			}

			// Aguardar um pouco antes de alternar new_name1
			time.Sleep(100 * time.Millisecond)

			// Atualizar new_name1 com o mesmo valor
			sharedState.mu.Lock()
			sharedState.currentNewName1 = newValue
			sharedState.mu.Unlock()

			payload.FlagName = "new_name1"
			if err := createFeatureFlag(serverURL, payload); err != nil {
				fmt.Printf("⚠️  Erro ao alternar new_name1 para %v: %v\n", newValue, err)
			} else {
				atomic.AddInt64(&stats.toggleCount, 1)
			}
		}
	}
}

func worker(id int, ff *featureflag.FeatureFlagSDK, stats *Stats, wg *sync.WaitGroup, stopChan chan bool, opsPerSecond int, sharedState *SharedState) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(opsPerSecond))
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			performOperations(ff, stats, sharedState)
		}
	}
}

func performOperations(ff *featureflag.FeatureFlagSDK, stats *Stats, sharedState *SharedState) {
	// Operação 1: GetFeatureFlag("invalid_ff").WithDefault(true)
	start := time.Now()
	isActive := ff.GetFeatureFlag("invalid_ff").WithDefault(true)
	recordOperation(stats, start, isActive, nil)

	// Operação 2: GetFeatureFlag("new_name").Err() com validação
	start = time.Now()
	isActive2, err := ff.GetFeatureFlag("new_name").Err()
	sharedState.mu.RLock()
	expectedNewName := sharedState.currentNewName
	sharedState.mu.RUnlock()

	recordOperationWithValidation(stats, start, isActive2, err, expectedNewName, "new_name")

	// Operação 3: GetFeatureFlag("new_name1").Val() com validação
	start = time.Now()
	isActive3 := ff.GetFeatureFlag("new_name1").Val()
	sharedState.mu.RLock()
	expectedNewName1 := sharedState.currentNewName1
	sharedState.mu.RUnlock()

	recordOperationWithValidation(stats, start, isActive3, nil, expectedNewName1, "new_name1")
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

func recordOperationWithValidation(stats *Stats, start time.Time, result bool, err error, expected bool, flagName string) {
	latency := time.Since(start).Microseconds()

	atomic.AddInt64(&stats.totalOperations, 1)
	atomic.AddInt64(&stats.totalLatency, latency)

	if err != nil {
		atomic.AddInt64(&stats.errorOperations, 1)
	} else {
		atomic.AddInt64(&stats.successOperations, 1)

		// Validar se o valor retornado está correto
		if result == expected {
			atomic.AddInt64(&stats.correctValues, 1)
		} else {
			atomic.AddInt64(&stats.validationErrors, 1)
			// Log de validação falhou (comentado para não poluir output)
			// fmt.Printf("⚠️  Validação falhou para %s: esperado=%v, recebido=%v\n", flagName, expected, result)
		}
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
		toggles := atomic.LoadInt64(&stats.toggleCount)
		validationErrs := atomic.LoadInt64(&stats.validationErrors)
		correctVals := atomic.LoadInt64(&stats.correctValues)

		rate := float64(total) / elapsed.Seconds()
		successRate := float64(0)
		if total > 0 {
			successRate = float64(success) / float64(total) * 100
		}

		validationRate := float64(0)
		if correctVals+validationErrs > 0 {
			validationRate = float64(correctVals) / float64(correctVals+validationErrs) * 100
		}

		fmt.Printf("⚡ [%5.1fs] Ops: %7d | Rate: %7.1f/s | Success: %6.2f%% | Validation: %6.2f%% | Toggles: %d | ValErrs: %d\n",
			elapsed.Seconds(),
			total,
			rate,
			successRate,
			validationRate,
			toggles,
			validationErrs,
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
	toggles := atomic.LoadInt64(&stats.toggleCount)
	validationErrs := atomic.LoadInt64(&stats.validationErrors)
	correctVals := atomic.LoadInt64(&stats.correctValues)

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
	fmt.Printf("   Ciclos completos: %d (3 ops/ciclo)\n", total/3)
	fmt.Printf("   Toggles realizados: %d\n", toggles)
	fmt.Println()

	fmt.Printf("⏱️  Latências:\n")
	fmt.Printf("   Mínima: %.3fms\n", float64(minLat)/1000.0)
	fmt.Printf("   Média: %.3fms\n", float64(avgLatency)/1000.0)
	fmt.Printf("   Máxima: %.3fms\n", float64(maxLat)/1000.0)
	fmt.Println()

	validationRate := float64(0)
	totalValidations := correctVals + validationErrs
	if totalValidations > 0 {
		validationRate = float64(correctVals) / float64(totalValidations) * 100
	}

	fmt.Printf("✅ Resultados:\n")
	fmt.Printf("   Sucesso: %d (%.2f%%)\n", success, successRate)
	fmt.Printf("   Erros: %d (%.2f%%)\n", errors, 100-successRate)
	fmt.Println()

	fmt.Printf("🔍 Validação de Estados:\n")
	fmt.Printf("   Valores corretos: %d (%.2f%%)\n", correctVals, validationRate)
	fmt.Printf("   Erros de validação: %d (%.2f%%)\n", validationErrs, 100-validationRate)
	fmt.Printf("   Total validado: %d (apenas new_name e new_name1)\n", totalValidations)
	fmt.Println()

	if validationRate >= 99.0 {
		fmt.Println("🎉 EXCELENTE! Validação de estados >= 99%")
	} else if validationRate >= 95.0 {
		fmt.Println("✅ BOM! Validação de estados >= 95%")
	} else if validationRate >= 90.0 {
		fmt.Println("⚠️  ACEITÁVEL! Validação de estados >= 90%")
	} else {
		fmt.Printf("❌ CRÍTICO! Validação de estados muito baixa: %.2f%%\n", validationRate)
		fmt.Println("   Isso indica que o SDK não está sincronizando corretamente com os toggles")
	}

	fmt.Println()
	fmt.Printf("🔄 Informação sobre toggles:\n")
	fmt.Printf("   Alterações de estado executadas: %d\n", toggles)
	fmt.Printf("   Flags alteradas a cada 2 segundos durante o teste\n")
	fmt.Println()
	fmt.Println("🎯 Teste de carga com toggle concluído!")
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
