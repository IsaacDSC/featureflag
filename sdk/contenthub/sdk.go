package contenthub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ContenthubSDK struct {
	host      string
	client    *http.Client
	ffDefault Value

	sleeper time.Duration
	db      map[string]Content
}

func NewContenthubSDK(hostFF string) *ContenthubSDK {
	sdk := &ContenthubSDK{client: &http.Client{}, host: hostFF, sleeper: time.Second * 60}
	return sdk
}

func (c *ContenthubSDK) WithEventualConsistency(time time.Duration) *ContenthubSDK {
	c.sleeper = time
	return c
}

// Listenner carrega os conteúdos na inicialização e mantém o cache em memória
// atualizado via polling periódico (WithEventualConsistency define o intervalo).
// A chamada bloqueia até o contexto ser cancelado, então normalmente é executada
// em uma goroutine.
func (c *ContenthubSDK) Listenner(ctx context.Context) (*ContenthubSDK, error) {
	contents, err := c.getAllContents(ctx)
	if err != nil {
		return nil, err
	}

	c.db = contents

	// Verificar se o servidor está acessível antes de iniciar o polling.
	clientWithTimeout := &http.Client{Timeout: 5 * time.Second}
	if _, err := clientWithTimeout.Get(c.host); err != nil {
		fmt.Println("❌ Servidor não está rodando!")
		fmt.Println("💡 Inicie o servidor primeiro: go run ./cmd")
		return nil, err
	}

	// Polling bloqueante até o contexto ser cancelado.
	c.refresh(ctx)

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

func (c *ContenthubSDK) Content(key string, sessionID ...string) Result {
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

func (c *ContenthubSDK) getAllContents(ctx context.Context) (map[string]Content, error) {
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

func (c *ContenthubSDK) refresh(ctx context.Context) {
	ticker := time.NewTicker(c.sleeper)
	defer ticker.Stop()

	fmt.Println("🔄 Refresh iniciado - atualizando contents a cada 60 segundos")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Refresh encerrado")
			return
		case <-ticker.C:
			contents, err := c.getAllContents(ctx)
			if err != nil {
				fmt.Printf("error on refresh contents: %v\n", err)
				continue
			}

			c.db = contents
			fmt.Println("✅ Contents atualizados via refresh")
		}
	}
}
