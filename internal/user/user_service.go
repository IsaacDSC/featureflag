package user

import (
	"context"
	"time"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository Adapter
}

func NewUserService(repository Adapter) *Service {
	return &Service{repository: repository}
}

// Register cria um novo usuário a partir de uma senha em texto puro,
// recebida sobre uma sessão já autenticada (POST /users). O hash é
// calculado aqui, nunca pelo chamador.
func (s Service) Register(ctx context.Context, emailInput, password string) error {
	if _, err := s.repository.GetByEmail(ctx, emailInput); err == nil {
		return errorutils.NewConflictError("user")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repository.Create(ctx, Entity{
		ID:           uuid.New(),
		Email:        emailInput,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	})
}

// Authenticate compara a senha em texto puro contra o hash persistido.
// Nunca revela se o email existe: qualquer falha (email não encontrado ou
// senha incorreta) retorna false.
func (s Service) Authenticate(ctx context.Context, emailInput, password string) bool {
	u, err := s.repository.GetByEmail(ctx, emailInput)
	if err != nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (s Service) Remove(ctx context.Context, emailInput string) error {
	return s.repository.Delete(ctx, emailInput)
}

func (s Service) List(ctx context.Context) ([]Entity, error) {
	return s.repository.ListAll(ctx)
}

// Seed insere um hash já pronto (vindo de ADMIN_USERS), sem passar por
// GenerateFromPassword — usado só pelo bootstrap do primeiro usuário, nunca
// pela API de criação via POST /users.
func (s Service) Seed(ctx context.Context, emailInput, passwordHashInput string) error {
	return s.repository.Create(ctx, Entity{
		ID:           uuid.New(),
		Email:        emailInput,
		PasswordHash: passwordHashInput,
		CreatedAt:    time.Now(),
	})
}
