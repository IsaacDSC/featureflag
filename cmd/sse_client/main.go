package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type FlagUpdateMessage struct {
	Type     string `json:"type"`
	FlagName string `json:"flagName"`
	NewValue bool   `json:"newValue"`
}

type FeatureFlag struct {
	ID        string `json:"id"`
	FlagName  string `json:"flag_name"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

func main() {
	serverURL := "http://localhost:3000/events/featureflag"

	fmt.Println("🔌 Conectando ao servidor SSE...")
	fmt.Printf("📍 URL: %s\n", serverURL)

	// Verificar se servidor está disponível
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get("http://localhost:3000")
	if err != nil {
		fmt.Println("❌ Servidor não está rodando!")
		fmt.Println("💡 Inicie o servidor primeiro: go run ./cmd/server")
		return
	}

	// Configurar context com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capturar sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Encerrando cliente...")
		cancel()
	}()

	// Conectar ao SSE
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		log.Fatal("Erro ao criar request:", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// Cliente sem timeout para SSE
	sseClient := &http.Client{}
	resp, err := sseClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Status inválido: %d", resp.StatusCode)
	}

	fmt.Println("✅ Conectado! Aguardando notificações...")
	fmt.Println("💡 Teste em outro terminal:")

	fmt.Println("   curl \"http://localhost:3000/update?flag=feature-a&value=true\"")
	fmt.Println("   curl \"http://localhost:3000/update?flag=feature-b&value=false\"")
	fmt.Println("🛑 Pressione Ctrl+C para sair")
	fmt.Println()

	scanner := bufio.NewScanner(resp.Body)

	// Canal para comunicação entre goroutines
	done := make(chan bool)

	// Goroutine para ler mensagens SSE
	go func() {
		defer func() { done <- true }()

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Tentar parsear como FlagUpdateMessage primeiro
				var msg FlagUpdateMessage
				if err := json.Unmarshal([]byte(data), &msg); err == nil && msg.Type == "updated-flag" {
					fmt.Printf("🔄 Flag atualizada: %s = %v\n", msg.FlagName, msg.NewValue)
					continue
				}

				// Tentar parsear como FeatureFlag
				var flag FeatureFlag
				if err := json.Unmarshal([]byte(data), &flag); err == nil {
					fmt.Printf("📦 Feature Flag recebida:\n")
					fmt.Printf("   ID: %s\n", flag.ID)
					fmt.Printf("   Nome: %s\n", flag.FlagName)
					fmt.Printf("   Ativa: %v\n", flag.Active)
					fmt.Printf("   Criada em: %s\n", flag.CreatedAt)
					continue
				}

				// Se não conseguir parsear, mostrar raw data
				fmt.Printf("📨 Mensagem recebida:\n")
				fmt.Printf("   %s\n", data)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("❌ Erro ao ler stream: %v\n", err)
		}
	}()

	// Aguardar cancelamento ou fim da leitura
	select {
	case <-ctx.Done():
		fmt.Println("🔌 Conexão cancelada pelo usuário")
	case <-done:
		fmt.Println("🔌 Conexão encerrada")
	}
}
