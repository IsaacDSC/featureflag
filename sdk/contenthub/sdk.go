package contenthub

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ContenthubSDK struct {
	host      string
	client    *http.Client
	ffDefault Value

	db map[string]Content
}

func NewContenthubSDK(hostFF string) *ContenthubSDK {
	sdk := &ContenthubSDK{client: &http.Client{}, host: hostFF}
	return sdk
}

func (c *ContenthubSDK) Listenner(ctx context.Context) (*ContenthubSDK, error) {
	contents, err := c.getAllContents(ctx)
	if err != nil {
		return nil, err
	}

	c.db = contents

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

	//FICAR ESCUTANDO EVENTOS (SSE)
	// Cliente com timeout apenas para verificar se o servidor está rodando
	clientWithTimeout := &http.Client{Timeout: 5 * time.Second}
	_, err = clientWithTimeout.Get(c.host)
	if err != nil {
		fmt.Println("❌ Servidor não está rodando!")
		fmt.Println("💡 Inicie o servidor primeiro: go run ./cmd/server")
		return nil, err
	}

	// Cliente sem timeout para a conexão SSE (que precisa ficar aberta)
	sseClient := &http.Client{}

	serverUrl := fmt.Sprintf("%s/events/contenthub", c.host)
	req, err := http.NewRequestWithContext(ctx, "GET", serverUrl, nil)
	if err != nil {
		log.Fatal("Erro ao criar request:", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := sseClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Status inválido: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())

			if data, ok := strings.CutPrefix(line, "data: "); ok {

				var content Content
				if err := json.Unmarshal([]byte(data), &content); err == nil {
					fmt.Printf("📦 Feature Content recebida:\n")
					fmt.Println()
					fmt.Printf("%+v\n", content)
					fmt.Println()

					c.db[content.Key] = content
					for _, v := range c.db {
						fmt.Printf("%+v\n", v)
					}
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

	select {
	case <-ctx.Done():
		fmt.Println("🔌 Conexão cancelada pelo usuário")
	case <-done:
		fmt.Println("🔌 Conexão encerrada")
	}

	return c, nil
}

type Result struct {
	value Value
	error error
}

func (fr Result) Err() (Value, error) {
	return fr.value, fr.error
}

func (fr Result) Val() Value {
	return fr.value
}

func (fr Result) String() string {
	return string(fr.value)
}

func (fr Result) DecodeJson(value any) error {
	return json.Unmarshal(fr.value, value)
}

func (c ContenthubSDK) Content(key string, sessionID ...string) Result {
	content, ok := c.db[key]

	if !ok {
		return Result{c.ffDefault, ErrNotFoundContenthub}
	}

	if len(sessionID) == 0 {
		return Result{content.Value(), nil}
	}

	ch := content.SessionStrategy.Val(sessionID[0])
	b, _ := json.Marshal(ch)

	return Result{b, nil}

}

func (c ContenthubSDK) getAllContents(ctx context.Context) (map[string]Content, error) {
	resp, err := http.Get(fmt.Sprintf("%s/contenthubs", c.host))
	if err != nil {
		return nil, fmt.Errorf("error on get features Contents :%w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error on io read all, %w", err)
	}

	var contents []Content
	if err := json.Unmarshal(body, &contents); err != nil {
		return nil, fmt.Errorf("error on json unmarshal, %w", err)
	}

	output := make(map[string]Content)
	for _, content := range contents {
		output[content.Key] = content
	}

	return output, nil
}
