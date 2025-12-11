package featureflag

import (
	"context"
	"testing"
	"time"

	"github.com/IsaacDSC/featureflag/internal/strategy"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB cria uma conexão de teste com MongoDB
// Retorna nil se o MongoDB não estiver disponível
func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use uma URI de teste - ajuste conforme necessário
	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skip("MongoDB não disponível:", err)
		return nil, nil
	}

	// Verificar se consegue fazer ping
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		t.Skip("MongoDB não está respondendo ou requer autenticação:", err)
		return nil, nil
	}

	// Usar um banco de dados de teste único
	dbName := "featureflags_test_" + uuid.New().String()
	db := client.Database(dbName)

	// Testar se consegue criar uma coleção (verificar permissões)
	testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer testCancel()

	testCollection := db.Collection("_test")
	_, err = testCollection.InsertOne(testCtx, map[string]string{"test": "test"})
	if err != nil {
		client.Disconnect(context.Background())
		t.Skip("MongoDB requer autenticação ou sem permissões:", err)
		return nil, nil
	}
	_ = testCollection.Drop(testCtx)

	// Função de cleanup
	cleanup := func() {
		ctx := context.Background()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	}

	return db, cleanup
}

func TestMongoDBRepository_SaveFF(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	entity := Entity{
		ID:        uuid.New(),
		FlagName:  "test-flag",
		Active:    true,
		CreatedAt: time.Now(),
		Strategies: strategy.Strategy{
			WithStrategy: true,
			SessionsID:   map[string]bool{"session-1": true},
			QtdCall:      0,
			Percent:      50.0,
		},
	}

	err := repo.SaveFF(entity)
	if err != nil {
		t.Fatalf("SaveFF falhou: %v", err)
	}

	// Verificar se foi salvo
	saved, err := repo.GetFF("test-flag")
	if err != nil {
		t.Fatalf("GetFF falhou após SaveFF: %v", err)
	}

	if saved.FlagName != entity.FlagName {
		t.Errorf("Esperado FlagName %s, obteve %s", entity.FlagName, saved.FlagName)
	}

	if saved.Active != entity.Active {
		t.Errorf("Esperado Active %v, obteve %v", entity.Active, saved.Active)
	}
}

func TestMongoDBRepository_SaveFF_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	entity := Entity{
		ID:        uuid.New(),
		FlagName:  "update-flag",
		Active:    false,
		CreatedAt: time.Now(),
	}

	// Salvar primeira vez
	err := repo.SaveFF(entity)
	if err != nil {
		t.Fatalf("Primeira SaveFF falhou: %v", err)
	}

	// Atualizar
	entity.Active = true
	err = repo.SaveFF(entity)
	if err != nil {
		t.Fatalf("Segunda SaveFF (update) falhou: %v", err)
	}

	// Verificar atualização
	updated, err := repo.GetFF("update-flag")
	if err != nil {
		t.Fatalf("GetFF falhou: %v", err)
	}

	if !updated.Active {
		t.Error("Esperado Active=true após update, obteve false")
	}
}

func TestMongoDBRepository_GetFF(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	entity := Entity{
		ID:        uuid.New(),
		FlagName:  "get-flag",
		Active:    true,
		CreatedAt: time.Now(),
	}

	err := repo.SaveFF(entity)
	if err != nil {
		t.Fatalf("SaveFF falhou: %v", err)
	}

	// Buscar existente
	found, err := repo.GetFF("get-flag")
	if err != nil {
		t.Fatalf("GetFF falhou: %v", err)
	}

	if found.FlagName != "get-flag" {
		t.Errorf("Esperado FlagName 'get-flag', obteve '%s'", found.FlagName)
	}
}

func TestMongoDBRepository_GetFF_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	// Tentar buscar flag inexistente
	_, err := repo.GetFF("non-existent")
	if err == nil {
		t.Fatal("Esperado erro ao buscar flag inexistente, obteve nil")
	}

	// Verificar se é um NotFoundError
	// A mensagem de erro deve conter "ff" ou similar
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Esperado mensagem de erro, obteve string vazia")
	}
}

func TestMongoDBRepository_GetAllFF(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	// Salvar múltiplas flags
	flags := []Entity{
		{
			ID:        uuid.New(),
			FlagName:  "flag-1",
			Active:    true,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			FlagName:  "flag-2",
			Active:    false,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			FlagName:  "flag-3",
			Active:    true,
			CreatedAt: time.Now(),
		},
	}

	for _, flag := range flags {
		err := repo.SaveFF(flag)
		if err != nil {
			t.Fatalf("SaveFF falhou para %s: %v", flag.FlagName, err)
		}
	}

	// Buscar todas
	all, err := repo.GetAllFF()
	if err != nil {
		t.Fatalf("GetAllFF falhou: %v", err)
	}

	if len(all) != len(flags) {
		t.Errorf("Esperado %d flags, obteve %d", len(flags), len(all))
	}

	// Verificar se todas as flags estão presentes
	for _, flag := range flags {
		if _, exists := all[flag.FlagName]; !exists {
			t.Errorf("Flag %s não encontrada no resultado GetAllFF", flag.FlagName)
		}
	}
}

func TestMongoDBRepository_GetAllFF_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	// Buscar em coleção vazia
	all, err := repo.GetAllFF()
	if err != nil {
		t.Fatalf("GetAllFF falhou em coleção vazia: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("Esperado 0 flags em coleção vazia, obteve %d", len(all))
	}
}

func TestMongoDBRepository_DeleteFF(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	entity := Entity{
		ID:        uuid.New(),
		FlagName:  "delete-flag",
		Active:    true,
		CreatedAt: time.Now(),
	}

	// Salvar
	err := repo.SaveFF(entity)
	if err != nil {
		t.Fatalf("SaveFF falhou: %v", err)
	}

	// Deletar
	err = repo.DeleteFF("delete-flag")
	if err != nil {
		t.Fatalf("DeleteFF falhou: %v", err)
	}

	// Verificar se foi deletado
	_, err = repo.GetFF("delete-flag")
	if err == nil {
		t.Fatal("Esperado erro ao buscar flag deletada, obteve nil")
	}
}

func TestMongoDBRepository_DeleteFF_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	// Tentar deletar flag inexistente
	err := repo.DeleteFF("non-existent")
	if err == nil {
		t.Fatal("Esperado erro ao deletar flag inexistente, obteve nil")
	}
}

func TestMongoDBRepository_InterfaceCompliance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	// Verificar se implementa a interface Adapter
	var _ Adapter = repo
}

func TestMongoDBRepository_ConcurrentOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	// Testar operações concurrent
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			entity := Entity{
				ID:        uuid.New(),
				FlagName:  "concurrent-flag",
				Active:    index%2 == 0,
				CreatedAt: time.Now(),
			}

			if err := repo.SaveFF(entity); err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Aguardar todas as goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	close(errors)

	// Verificar erros
	for err := range errors {
		t.Errorf("Erro em operação concurrent: %v", err)
	}

	// Verificar se a flag existe (deve ter sido salva pelo menos uma vez)
	_, err := repo.GetFF("concurrent-flag")
	if err != nil {
		t.Fatalf("GetFF falhou após operações concurrent: %v", err)
	}
}

func BenchmarkMongoDBRepository_SaveFF(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		b.Skip("MongoDB não disponível para benchmark:", err)
		return
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		b.Skip("MongoDB não está respondendo para benchmark:", err)
		return
	}

	db := client.Database("featureflags_bench")
	defer func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	entity := Entity{
		ID:        uuid.New(),
		FlagName:  "bench-flag",
		Active:    true,
		CreatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.SaveFF(entity)
	}
}

func BenchmarkMongoDBRepository_GetFF(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		b.Skip("MongoDB não disponível para benchmark:", err)
		return
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		b.Skip("MongoDB não está respondendo para benchmark:", err)
		return
	}

	db := client.Database("featureflags_bench")
	defer func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}()

	repo := NewMongoDBFeatureFlagRepository(db, "flags")

	entity := Entity{
		ID:        uuid.New(),
		FlagName:  "bench-flag",
		Active:    true,
		CreatedAt: time.Now(),
	}

	_ = repo.SaveFF(entity)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetFF("bench-flag")
	}
}
