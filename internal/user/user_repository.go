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
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"password_hash"`
	Role               string    `json:"role"`
	Projects           []string  `json:"projects"`
	CreatedAt          time.Time `json:"created_at"`
	MustChangePassword bool      `json:"must_change_password"`
}

func toRecord(e Entity) record {
	return record{
		ID:                 e.ID,
		Email:              e.Email,
		PasswordHash:       e.PasswordHash,
		Role:               e.Role,
		Projects:           e.Projects,
		CreatedAt:          e.CreatedAt,
		MustChangePassword: e.MustChangePassword,
	}
}

func (r record) toEntity() Entity {
	return Entity{
		ID:                 r.ID,
		Email:              r.Email,
		PasswordHash:       r.PasswordHash,
		Role:               r.Role,
		Projects:           r.Projects,
		CreatedAt:          r.CreatedAt,
		MustChangePassword: r.MustChangePassword,
	}
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

// Update altera só role/projects de um registro já existente — preserva
// password_hash, id e created_at do que já estava salvo.
func (r Repository) Update(ctx context.Context, input Entity) error {
	store, err := r.readStore()
	if err != nil {
		return err
	}

	existing, ok := store[input.Email]
	if !ok {
		return errorutils.NewNotFoundError("user")
	}

	existing.Role = input.Role
	existing.Projects = input.Projects
	store[input.Email] = existing

	return r.writeStore(store)
}

// UpdatePassword grava um novo hash de senha e limpa MustChangePassword —
// preserva role, projects, id e created_at do que já estava salvo.
func (r Repository) UpdatePassword(ctx context.Context, emailInput, passwordHashInput string) error {
	store, err := r.readStore()
	if err != nil {
		return err
	}

	existing, ok := store[emailInput]
	if !ok {
		return errorutils.NewNotFoundError("user")
	}

	existing.PasswordHash = passwordHashInput
	existing.MustChangePassword = false
	store[emailInput] = existing

	return r.writeStore(store)
}

// RequirePasswordChange marca um usuário já existente para trocar a senha
// no próximo login — preserva password_hash, role, projects, id e
// created_at do que já estava salvo.
func (r Repository) RequirePasswordChange(ctx context.Context, emailInput string) error {
	store, err := r.readStore()
	if err != nil {
		return err
	}

	existing, ok := store[emailInput]
	if !ok {
		return errorutils.NewNotFoundError("user")
	}

	existing.MustChangePassword = true
	store[emailInput] = existing

	return r.writeStore(store)
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
