package user

import (
	"context"
	"testing"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type fakeRepository struct {
	byEmail map[string]Entity
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byEmail: map[string]Entity{}}
}

func (r *fakeRepository) Create(ctx context.Context, input Entity) error {
	r.byEmail[input.Email] = input
	return nil
}

func (r *fakeRepository) GetByEmail(ctx context.Context, emailInput string) (Entity, error) {
	if u, ok := r.byEmail[emailInput]; ok {
		return u, nil
	}
	return Entity{}, errorutils.NewNotFoundError("user")
}

func (r *fakeRepository) Delete(ctx context.Context, emailInput string) error {
	delete(r.byEmail, emailInput)
	return nil
}

func (r *fakeRepository) ListAll(ctx context.Context) ([]Entity, error) {
	result := make([]Entity, 0, len(r.byEmail))
	for _, u := range r.byEmail {
		result = append(result, u)
	}
	return result, nil
}

func TestService_Register(t *testing.T) {
	t.Run("creates a new user with a bcrypt hash, never the plaintext password", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)

		if err := svc.Register(context.Background(), "alice@example.com", "s3cret"); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		stored, err := repo.GetByEmail(context.Background(), "alice@example.com")
		if err != nil {
			t.Fatalf("GetByEmail() unexpected error = %v", err)
		}
		if stored.PasswordHash == "s3cret" || stored.PasswordHash == "" {
			t.Errorf("PasswordHash = %q, want a bcrypt hash, not the plaintext password", stored.PasswordHash)
		}
	})

	t.Run("duplicate email returns a conflict error", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)

		if err := svc.Register(context.Background(), "alice@example.com", "s3cret"); err != nil {
			t.Fatalf("first Register() unexpected error = %v", err)
		}

		err := svc.Register(context.Background(), "alice@example.com", "other-password")
		if _, ok := err.(*errorutils.ConflictError); !ok {
			t.Errorf("Register() error = %v, want *errorutils.ConflictError", err)
		}
	})
}

func TestService_Authenticate(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)
	if err := svc.Register(context.Background(), "alice@example.com", "s3cret"); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
		want     bool
	}{
		{name: "correct password", email: "alice@example.com", password: "s3cret", want: true},
		{name: "wrong password", email: "alice@example.com", password: "wrong", want: false},
		{name: "unknown email", email: "bob@example.com", password: "s3cret", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.Authenticate(context.Background(), tt.email, tt.password); got != tt.want {
				t.Errorf("Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Seed(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	// Seed grava o hash tal como veio (sem GenerateFromPassword) — simula o
	// bootstrap via ADMIN_USERS, onde o hash já foi gerado offline.
	precomputedHash := "$2a$10$abcdefghijklmnopqrstuv"
	if err := svc.Seed(context.Background(), "admin@example.com", precomputedHash); err != nil {
		t.Fatalf("Seed() unexpected error = %v", err)
	}

	stored, err := repo.GetByEmail(context.Background(), "admin@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() unexpected error = %v", err)
	}
	if stored.PasswordHash != precomputedHash {
		t.Errorf("PasswordHash = %q, want %q (Seed must not re-hash)", stored.PasswordHash, precomputedHash)
	}
}
