package contenthub

import (
	"context"
	"testing"
	"time"

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
	dbName := "contenthub_test_" + uuid.New().String()
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

func TestMongoDBRepository_SaveContentHub(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "test-variable",
		Value:       "test-value",
		Description: "test description",
		Active:      true,
		CreatedAt:   time.Now(),
		SessionsStrategies: SessionsStrategies{
			{SessionID: "default", Response: "default-response"},
		},
		BalancerStrategy: BalancerStrategy{
			{Weight: 50, Response: "response-a"},
			{Weight: 50, Response: "response-b"},
		},
	}

	err = repo.SaveContentHub(entity)
	if err != nil {
		t.Fatalf("SaveContentHub falhou: %v", err)
	}

	// Verificar se foi salvo
	saved, err := repo.GetContentHub("test-variable")
	if err != nil {
		t.Fatalf("GetContentHub falhou após SaveContentHub: %v", err)
	}

	if saved.Variable != entity.Variable {
		t.Errorf("Esperado Variable %s, obteve %s", entity.Variable, saved.Variable)
	}

	if saved.Value != entity.Value {
		t.Errorf("Esperado Value %s, obteve %s", entity.Value, saved.Value)
	}

	if saved.Active != entity.Active {
		t.Errorf("Esperado Active %v, obteve %v", entity.Active, saved.Active)
	}
}

func TestMongoDBRepository_SaveContentHub_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "update-variable",
		Value:       "initial-value",
		Description: "initial description",
		Active:      false,
		CreatedAt:   time.Now(),
	}

	// Salvar primeira vez
	err = repo.SaveContentHub(entity)
	if err != nil {
		t.Fatalf("Primeira SaveContentHub falhou: %v", err)
	}

	// Atualizar
	entity.Value = "updated-value"
	entity.Active = true
	err = repo.SaveContentHub(entity)
	if err != nil {
		t.Fatalf("Segunda SaveContentHub (update) falhou: %v", err)
	}

	// Verificar atualização
	updated, err := repo.GetContentHub("update-variable")
	if err != nil {
		t.Fatalf("GetContentHub falhou: %v", err)
	}

	if updated.Value != "updated-value" {
		t.Errorf("Esperado Value 'updated-value', obteve '%s'", updated.Value)
	}

	if !updated.Active {
		t.Error("Esperado Active=true após update, obteve false")
	}
}

func TestMongoDBRepository_GetContentHub(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "get-variable",
		Value:       "get-value",
		Description: "get description",
		Active:      true,
		CreatedAt:   time.Now(),
	}

	err = repo.SaveContentHub(entity)
	if err != nil {
		t.Fatalf("SaveContentHub falhou: %v", err)
	}

	// Buscar existente
	found, err := repo.GetContentHub("get-variable")
	if err != nil {
		t.Fatalf("GetContentHub falhou: %v", err)
	}

	if found.Variable != "get-variable" {
		t.Errorf("Esperado Variable 'get-variable', obteve '%s'", found.Variable)
	}

	if found.Value != "get-value" {
		t.Errorf("Esperado Value 'get-value', obteve '%s'", found.Value)
	}
}

func TestMongoDBRepository_GetContentHub_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Tentar buscar variable inexistente
	_, err = repo.GetContentHub("non-existent")
	if err == nil {
		t.Fatal("Esperado erro ao buscar variable inexistente, obteve nil")
	}

	// Verificar se é um NotFoundError
	// A mensagem de erro deve conter "ff" ou similar
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Esperado mensagem de erro, obteve string vazia")
	}
}

func TestMongoDBRepository_GetAllContentHub(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Salvar múltiplas entidades
	entities := []Entity{
		{
			ID:          uuid.New(),
			Variable:    "variable-1",
			Value:       "value-1",
			Description: "description 1",
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Variable:    "variable-2",
			Value:       "value-2",
			Description: "description 2",
			Active:      false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Variable:    "variable-3",
			Value:       "value-3",
			Description: "description 3",
			Active:      true,
			CreatedAt:   time.Now(),
		},
	}

	for _, entity := range entities {
		err := repo.SaveContentHub(entity)
		if err != nil {
			t.Fatalf("SaveContentHub falhou para %s: %v", entity.Variable, err)
		}
	}

	// Buscar todas
	all, err := repo.GetAllContentHub()
	if err != nil {
		t.Fatalf("GetAllContentHub falhou: %v", err)
	}

	if len(all) != len(entities) {
		t.Errorf("Esperado %d entidades, obteve %d", len(entities), len(all))
	}

	// Verificar se todas as entidades estão presentes
	for _, entity := range entities {
		if _, exists := all[entity.Variable]; !exists {
			t.Errorf("Variable %s não encontrada no resultado GetAllContentHub", entity.Variable)
		}
	}
}

func TestMongoDBRepository_GetAllContentHub_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Buscar em coleção vazia
	all, err := repo.GetAllContentHub()
	if err != nil {
		t.Fatalf("GetAllContentHub falhou em coleção vazia: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("Esperado 0 entidades em coleção vazia, obteve %d", len(all))
	}
}

func TestMongoDBRepository_DeleteContentHub(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "delete-variable",
		Value:       "delete-value",
		Description: "delete description",
		Active:      true,
		CreatedAt:   time.Now(),
	}

	// Salvar
	err = repo.SaveContentHub(entity)
	if err != nil {
		t.Fatalf("SaveContentHub falhou: %v", err)
	}

	// Deletar
	err = repo.DeleteContentHub("delete-variable")
	if err != nil {
		t.Fatalf("DeleteContentHub falhou: %v", err)
	}

	// Verificar se foi deletado
	_, err = repo.GetContentHub("delete-variable")
	if err == nil {
		t.Fatal("Esperado erro ao buscar variable deletada, obteve nil")
	}
}

func TestMongoDBRepository_DeleteContentHub_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Tentar deletar variable inexistente
	err = repo.DeleteContentHub("non-existent")
	if err == nil {
		t.Fatal("Esperado erro ao deletar variable inexistente, obteve nil")
	}
}

func TestMongoDBRepository_InterfaceCompliance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Verificar se implementa a interface Adapter
	var _ Adapter = repo
}

func TestMongoDBRepository_ConcurrentOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Testar operações concurrent
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			entity := Entity{
				ID:          uuid.New(),
				Variable:    "concurrent-variable",
				Value:       "concurrent-value",
				Description: "concurrent description",
				Active:      index%2 == 0,
				CreatedAt:   time.Now(),
			}

			if err := repo.SaveContentHub(entity); err != nil {
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

	// Verificar se a variable existe (deve ter sido salva pelo menos uma vez)
	_, err = repo.GetContentHub("concurrent-variable")
	if err != nil {
		t.Fatalf("GetContentHub falhou após operações concurrent: %v", err)
	}
}

func TestMongoDBRepository_WithStrategies(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "strategy-variable",
		Value:       "strategy-value",
		Description: "test with strategies",
		Active:      true,
		CreatedAt:   time.Now(),
		SessionsStrategies: SessionsStrategies{
			{SessionID: "default", Response: "default-response"},
			{SessionID: "session-1", Response: "session-1-response"},
			{SessionID: "session-2", Response: "session-2-response"},
		},
		BalancerStrategy: BalancerStrategy{
			{Weight: 30, Response: "response-a", Qtt: 0},
			{Weight: 70, Response: "response-b", Qtt: 0},
		},
	}

	err = repo.SaveContentHub(entity)
	if err != nil {
		t.Fatalf("SaveContentHub com strategies falhou: %v", err)
	}

	// Verificar se as strategies foram salvas corretamente
	saved, err := repo.GetContentHub("strategy-variable")
	if err != nil {
		t.Fatalf("GetContentHub falhou: %v", err)
	}

	if len(saved.SessionsStrategies) != 3 {
		t.Errorf("Esperado 3 SessionsStrategies, obteve %d", len(saved.SessionsStrategies))
	}

	if len(saved.BalancerStrategy) != 2 {
		t.Errorf("Esperado 2 BalancerStrategy, obteve %d", len(saved.BalancerStrategy))
	}
}

func BenchmarkMongoDBRepository_SaveContentHub(b *testing.B) {
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

	db := client.Database("contenthub_bench")
	defer func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "bench-variable",
		Value:       "bench-value",
		Description: "benchmark description",
		Active:      true,
		CreatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.SaveContentHub(entity)
	}
}

func BenchmarkMongoDBRepository_GetContentHub(b *testing.B) {
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

	db := client.Database("contenthub_bench")
	defer func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}()

	repo, err := NewMongoDBContentHubRepository(db)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}

	entity := Entity{
		ID:          uuid.New(),
		Variable:    "bench-variable",
		Value:       "bench-value",
		Description: "benchmark description",
		Active:      true,
		CreatedAt:   time.Now(),
	}

	_ = repo.SaveContentHub(entity)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetContentHub("bench-variable")
	}
}
