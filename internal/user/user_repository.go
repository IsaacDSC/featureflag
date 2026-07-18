package user

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/google/uuid"
)

type Repository struct{}

func NewUserRepository() *Repository {
	return &Repository{}
}

// record é a representação em disco de um usuário. Entity.PasswordHash usa
// `json:"-"` de propósito para nunca vazar em respostas HTTP — mas isso
// também impediria o hash de ser persistido se o arquivo gravasse a Entity
// diretamente via encoding/json. record existe só para o jsonfile guardar
// o hash corretamente, sem afetar o contrato de serialização HTTP da Entity.
type record struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

func toRecord(e Entity) record {
	return record{ID: e.ID, Email: e.Email, PasswordHash: e.PasswordHash, CreatedAt: e.CreatedAt}
}

func (r record) toEntity() Entity {
	return Entity{ID: r.ID, Email: r.Email, PasswordHash: r.PasswordHash, CreatedAt: r.CreatedAt}
}

type userStore map[string]record // email -> record

func (r Repository) readStore() (userStore, error) {
	b, err := os.ReadFile(env.FilePathUsers)
	if err != nil {
		if os.IsNotExist(err) {
			return userStore{}, nil
		}
		return nil, err
	}

	if len(b) == 0 {
		return userStore{}, nil
	}

	var store userStore
	if err := json.Unmarshal(b, &store); err != nil {
		return nil, err
	}

	if store == nil {
		store = userStore{}
	}

	return store, nil
}

func (r Repository) writeStore(store userStore) error {
	b, err := json.Marshal(store)
	if err != nil {
		return err
	}

	return os.WriteFile(env.FilePathUsers, b, 0644)
}

func (r Repository) Create(ctx context.Context, input Entity) error {
	store, err := r.readStore()
	if err != nil {
		return err
	}

	store[input.Email] = toRecord(input)

	return r.writeStore(store)
}

func (r Repository) GetByEmail(ctx context.Context, emailInput string) (Entity, error) {
	store, err := r.readStore()
	if err != nil {
		return Entity{}, err
	}

	if output, ok := store[emailInput]; ok {
		return output.toEntity(), nil
	}

	return Entity{}, errorutils.NewNotFoundError("user")
}

func (r Repository) Delete(ctx context.Context, emailInput string) error {
	store, err := r.readStore()
	if err != nil {
		return err
	}

	delete(store, emailInput)

	return r.writeStore(store)
}

func (r Repository) ListAll(ctx context.Context) ([]Entity, error) {
	store, err := r.readStore()
	if err != nil {
		return nil, err
	}

	result := make([]Entity, 0, len(store))
	for _, rec := range store {
		result = append(result, rec.toEntity())
	}

	return result, nil
}
