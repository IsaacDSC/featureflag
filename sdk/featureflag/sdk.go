package featureflag

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

type FeatureFlagSDK struct {
	host      string
	client    *http.Client
	ffDefault bool

	sleeper       time.Duration
	inMemoryFlags map[string]Flag
}

func NewFeatureFlagSDK(hostFF string) *FeatureFlagSDK {
	sdk := &FeatureFlagSDK{client: &http.Client{}, host: hostFF, sleeper: time.Second * 60}
	return sdk
}

func (c *FeatureFlagSDK) WithEventualConsistency(time time.Duration) *FeatureFlagSDK {
	c.sleeper = time
	return c
}

func (ff *FeatureFlagSDK) Listenner(ctx context.Context) (*FeatureFlagSDK, error) {
	flags, err := ff.getAllFlags(ctx)
	if err != nil {
		return nil, err
	}

	ff.inMemoryFlags = flags

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

	// Iniciar refresh em background
	go ff.refresh(ctx)

	//FICAR ESCUTANDO EVENTOS (SSE)
	// Cliente com timeout apenas para verificar se o servidor está rodando
	clientWithTimeout := &http.Client{Timeout: 5 * time.Second}
	_, err = clientWithTimeout.Get(ff.host)
	if err != nil {
		fmt.Println("❌ Servidor não está rodando!")
		fmt.Println("💡 Inicie o servidor primeiro: go run ./cmd/server")
		return nil, err
	}

	// Cliente sem timeout para a conexão SSE (que precisa ficar aberta)
	sseClient := &http.Client{}

	serverUrl := fmt.Sprintf("%s/events/featureflag", ff.host)
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

				var flag Flag
				if err := json.Unmarshal([]byte(data), &flag); err == nil {
					fmt.Printf("📦 Feature Flag recebida:\n")
					fmt.Println()
					fmt.Printf("%+v\n", flag)
					fmt.Println()

					ff.inMemoryFlags[flag.FlagName] = flag
					for _, v := range ff.inMemoryFlags {
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

	return ff, nil
}

type FFResponse struct {
	Bool  bool
	Error error
}

func (fr FFResponse) WithDefault(ffDefault bool) bool {
	if fr.Error != nil {
		return ffDefault
	}

	return fr.Bool
}

func (fr FFResponse) Err() (bool, error) {
	return fr.Bool, fr.Error
}

func (fr FFResponse) Val() bool {
	return fr.Bool
}

func (ff FeatureFlagSDK) GetFeatureFlag(key string, sessionID ...string) FFResponse {
	flag, ok := ff.inMemoryFlags[key]

	if !ok {
		return FFResponse{ff.ffDefault, ErrNotFoundFeatureFlag}
	}

	usingStrategy := flag.IsUseStrategy()
	if !usingStrategy {
		return FFResponse{flag.Active, nil}
	}

	if len(sessionID) == 0 {
		return FFResponse{ff.ffDefault, ErrInvalidStrategy}
	}

	isActive := flag.ValidateStrategy(sessionID[0]).
		Increment().
		Bool()

	return FFResponse{isActive, nil}
}

func (ff FeatureFlagSDK) getAllFlags(ctx context.Context) (map[string]Flag, error) {
	resp, err := http.Get(fmt.Sprintf("%s/featureflags", ff.host))
	if err != nil {
		return nil, fmt.Errorf("error on get features flags :%w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error on io read all, %w", err)
	}

	var flags []Flag
	if err := json.Unmarshal(body, &flags); err != nil {
		return nil, fmt.Errorf("error on decode json: %w", err)
	}

	output := make(map[string]Flag)
	for _, flag := range flags {
		output[flag.FlagName] = flag
	}

	return output, nil
}

func (ff FeatureFlagSDK) refresh(ctx context.Context) {
	ticker := time.NewTicker(ff.sleeper)
	defer ticker.Stop()

	fmt.Println("🔄 Refresh iniciado - atualizando flags a cada 60 segundos")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Refresh encerrado")
			return
		case <-ticker.C:
			flags, err := ff.getAllFlags(ctx)
			if err != nil {
				fmt.Printf("error on refresh flags: %v\n", err)
				continue
			}

			ff.inMemoryFlags = flags
			fmt.Println("✅ Flags atualizadas via refresh")
		}
	}
}
