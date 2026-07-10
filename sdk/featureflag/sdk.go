package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

type FeatureFlagSDK struct {
	host      string
	project   string
	client    *http.Client
	ffDefault bool

	sleeper       time.Duration
	inMemoryFlags map[string]Flag
}

func NewFeatureFlagSDK(hostFF string, project string) *FeatureFlagSDK {
	sdk := &FeatureFlagSDK{client: &http.Client{}, host: hostFF, project: project, sleeper: time.Second * 60}
	return sdk
}

func (c *FeatureFlagSDK) WithEventualConsistency(time time.Duration) *FeatureFlagSDK {
	c.sleeper = time
	return c
}

// Listenner carrega as flags do project na inicialização e mantém o cache em
// memória atualizado via polling periódico (WithEventualConsistency define o
// intervalo). A chamada bloqueia até o contexto ser cancelado, então normalmente
// é executada em uma goroutine.
func (ff *FeatureFlagSDK) Listenner(ctx context.Context) (*FeatureFlagSDK, error) {
	flags, err := ff.getAllFlags(ctx)
	if err != nil {
		return nil, err
	}

	ff.inMemoryFlags = flags

	// Verificar se o servidor está acessível antes de iniciar o polling.
	clientWithTimeout := &http.Client{Timeout: 5 * time.Second}
	if _, err := clientWithTimeout.Get(ff.host); err != nil {
		fmt.Println("❌ Servidor não está rodando!")
		fmt.Println("💡 Inicie o servidor primeiro: go run ./cmd")
		return nil, err
	}

	// Polling bloqueante até o contexto ser cancelado.
	ff.refresh(ctx)

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

func (ff *FeatureFlagSDK) GetFeatureFlag(key string, sessionID ...string) FFResponse {
	flag, ok := ff.inMemoryFlags[key]

	if !ok {
		return FFResponse{ff.ffDefault, ErrNotFoundFeatureFlag}
	}

	usingStrategy := flag.IsUseStrategy()
	if !usingStrategy {
		return FFResponse{flag.Active, nil}
	}

	if len(sessionID) > 0 {
		updatedFlag := flag.ValidateStrategy(sessionID[0]).Increment()
		ff.inMemoryFlags[key] = updatedFlag
		return FFResponse{updatedFlag.Active, nil}
	}

	updatedFlag := flag.Balancer()
	ff.inMemoryFlags[key] = updatedFlag
	return FFResponse{updatedFlag.Active, nil}
}

func (ff FeatureFlagSDK) getAllFlags(ctx context.Context) (map[string]Flag, error) {
	resp, err := http.Get(fmt.Sprintf("%s/featureflag/%s/all", ff.host, ff.project))
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

// hasChanged compares two flags and returns true if there was a change in the relevant fields.
// Fields compared: Active, Strategy.SessionsID, Strategy.Percent, Strategy.WithStrategy
func hasChanged(oldFlag, newFlag Flag) bool {
	if oldFlag.Active != newFlag.Active {
		return true
	}

	if oldFlag.Strategy.WithStrategy != newFlag.Strategy.WithStrategy {
		return true
	}

	if oldFlag.Strategy.Percent != newFlag.Strategy.Percent {
		return true
	}

	if !reflect.DeepEqual(oldFlag.Strategy.SessionsID, newFlag.Strategy.SessionsID) {
		return true
	}

	return false
}

// filterChangedFlags compares server flags with in-memory flags
// and returns a map containing only the flags that were changed.
// Preserves the local state (QtdCall) of existing flags when necessary.
func filterChangedFlags(serverFlags, memoryFlags map[string]Flag) map[string]Flag {
	changedFlags := make(map[string]Flag)

	for flagName, serverFlag := range serverFlags {
		memoryFlag, existsInMemory := memoryFlags[flagName]

		// New flag: add directly
		if !existsInMemory {
			changedFlags[flagName] = serverFlag
			continue
		}

		// Existing flag: check if changed
		if hasChanged(memoryFlag, serverFlag) {
			// Preserve local QtdCall if strategy is still active
			if serverFlag.Strategy.WithStrategy && memoryFlag.Strategy.WithStrategy {
				serverFlag.Strategy.QtdCall = memoryFlag.Strategy.QtdCall
			}
			changedFlags[flagName] = serverFlag
		}
	}

	return changedFlags
}

// mergeFlags merges changed flags with in-memory flags,
// also removes flags that were deleted on the server.
func mergeFlags(memoryFlags, serverFlags, changedFlags map[string]Flag) map[string]Flag {
	result := make(map[string]Flag)

	// Iterate over server flags to ensure deleted flags are removed
	for flagName := range serverFlags {
		if changedFlag, wasChanged := changedFlags[flagName]; wasChanged {
			// Flag was changed: use the new version
			result[flagName] = changedFlag
		} else if memoryFlag, existsInMemory := memoryFlags[flagName]; existsInMemory {
			// Flag unchanged: keep in-memory version (preserves QtdCall)
			result[flagName] = memoryFlag
		}
	}

	return result
}

func (ff *FeatureFlagSDK) refresh(ctx context.Context) {
	ticker := time.NewTicker(ff.sleeper)
	defer ticker.Stop()

	fmt.Printf("🔄 Refresh started - updating flags every %v\n", ff.sleeper)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Refresh stopped")
			return
		case <-ticker.C:
			serverFlags, err := ff.getAllFlags(ctx)
			if err != nil {
				fmt.Printf("error on refresh flags: %v\n", err)
				continue
			}

			// Filter only the flags that changed
			changedFlags := filterChangedFlags(serverFlags, ff.inMemoryFlags)

			if len(changedFlags) > 0 {
				// Merge changed flags with in-memory flags
				ff.inMemoryFlags = mergeFlags(ff.inMemoryFlags, serverFlags, changedFlags)
				fmt.Printf("✅ %d flag(s) updated via refresh\n", len(changedFlags))
			} else {
				fmt.Println("ℹ️  No changes detected on refresh")
			}
		}
	}
}
