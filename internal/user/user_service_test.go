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

func (r *fakeRepository) Update(ctx context.Context, input Entity) error {
	existing, ok := r.byEmail[input.Email]
	if !ok {
		return errorutils.NewNotFoundError("user")
	}
	existing.Role = input.Role
	existing.Projects = input.Projects
	r.byEmail[input.Email] = existing
	return nil
}

func (r *fakeRepository) UpdatePassword(ctx context.Context, emailInput, passwordHashInput string) error {
	existing, ok := r.byEmail[emailInput]
	if !ok {
		return errorutils.NewNotFoundError("user")
	}
	existing.PasswordHash = passwordHashInput
	existing.MustChangePassword = false
	r.byEmail[emailInput] = existing
	return nil
}

func (r *fakeRepository) RequirePasswordChange(ctx context.Context, emailInput string) error {
	existing, ok := r.byEmail[emailInput]
	if !ok {
		return errorutils.NewNotFoundError("user")
	}
	existing.MustChangePassword = true
	r.byEmail[emailInput] = existing
	return nil
}

func TestService_Register(t *testing.T) {
	t.Run("creates a new user with a bcrypt hash, never the plaintext password", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)

		if err := svc.Register(context.Background(), "alice@example.com", "s3cret", RoleMember, []string{"checkout"}); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		stored, err := repo.GetByEmail(context.Background(), "alice@example.com")
		if err != nil {
			t.Fatalf("GetByEmail() unexpected error = %v", err)
		}
		if stored.PasswordHash == "s3cret" || stored.PasswordHash == "" {
			t.Errorf("PasswordHash = %q, want a bcrypt hash, not the plaintext password", stored.PasswordHash)
		}
		if stored.Role != RoleMember || len(stored.Projects) != 1 || stored.Projects[0] != "checkout" {
			t.Errorf("Role/Projects = %q/%v, want member/[checkout]", stored.Role, stored.Projects)
		}
	})

	t.Run("duplicate email returns a conflict error", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)

		if err := svc.Register(context.Background(), "alice@example.com", "s3cret", RoleMember, nil); err != nil {
			t.Fatalf("first Register() unexpected error = %v", err)
		}

		err := svc.Register(context.Background(), "alice@example.com", "other-password", RoleMember, nil)
		if _, ok := err.(*errorutils.ConflictError); !ok {
			t.Errorf("Register() error = %v, want *errorutils.ConflictError", err)
		}
	})
}

func TestService_Authenticate(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)
	if err := svc.Register(context.Background(), "alice@example.com", "s3cret", RoleMember, nil); err != nil {
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
	if stored.Role != RoleAdmin {
		t.Errorf("Role = %q, want %q (bootstrap user must be admin)", stored.Role, RoleAdmin)
	}
}

// TestService_LegacyEmptyRole é um teste de regressão: usuários persistidos
// antes de Role/Projects existirem no schema (documento sem o campo "role")
// devem continuar com o acesso irrestrito que já tinham, não ficar
// travados de fora por engano.
func TestService_LegacyEmptyRole(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	// Simula um registro legado, gravado direto no repositório (sem passar
	// por Register, que já exige um role válido) — reproduz o dado real
	// encontrado em produção antes desta correção.
	repo.byEmail["legacy@example.com"] = Entity{Email: "legacy@example.com", PasswordHash: "hash"}

	isAdmin, err := svc.IsAdmin(context.Background(), "legacy@example.com")
	if err != nil {
		t.Fatalf("IsAdmin() unexpected error = %v", err)
	}
	if !isAdmin {
		t.Error("IsAdmin() = false, want true for a legacy user with no role field")
	}

	hasAccess, err := svc.HasProjectAccess(context.Background(), "legacy@example.com", "any-project")
	if err != nil {
		t.Fatalf("HasProjectAccess() unexpected error = %v", err)
	}
	if !hasAccess {
		t.Error("HasProjectAccess() = false, want true for a legacy user with no role field")
	}

	got, err := svc.GetByEmail(context.Background(), "legacy@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() unexpected error = %v", err)
	}
	if got.Role != RoleAdmin {
		t.Errorf("GetByEmail().Role = %q, want %q", got.Role, RoleAdmin)
	}

	users, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() unexpected error = %v", err)
	}
	if len(users) != 1 || users[0].Role != RoleAdmin {
		t.Errorf("List() = %+v, want a single user normalized to role %q", users, RoleAdmin)
	}
}

func TestService_HasProjectAccess(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	if err := svc.Register(context.Background(), "admin@example.com", "pw", RoleAdmin, nil); err != nil {
		t.Fatalf("Register(admin) unexpected error = %v", err)
	}
	if err := svc.Register(context.Background(), "member@example.com", "pw", RoleMember, []string{"checkout"}); err != nil {
		t.Fatalf("Register(member) unexpected error = %v", err)
	}

	tests := []struct {
		name    string
		email   string
		project string
		want    bool
	}{
		{name: "admin has access to any project", email: "admin@example.com", project: "billing", want: true},
		{name: "member has access to a granted project", email: "member@example.com", project: "checkout", want: true},
		{name: "member has no access to an ungranted project", email: "member@example.com", project: "billing", want: false},
		{name: "unknown email has no access", email: "ghost@example.com", project: "checkout", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.HasProjectAccess(context.Background(), tt.email, tt.project)
			if err != nil {
				t.Fatalf("HasProjectAccess() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("HasProjectAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestService_Register_MustChangePassword é um teste de regressão: uma
// conta criada por um admin (senha temporária) deve exigir troca de senha no
// primeiro login.
func TestService_Register_MustChangePassword(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	if err := svc.Register(context.Background(), "alice@example.com", "s3cret", RoleMember, nil); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	must, err := svc.MustChangePassword(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("MustChangePassword() unexpected error = %v", err)
	}
	if !must {
		t.Error("MustChangePassword() = false, want true right after Register()")
	}
}

// TestService_Seed_MustChangePassword covers the bootstrap admin path: the
// very first login must also force a password change.
func TestService_Seed_MustChangePassword(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	if err := svc.Seed(context.Background(), "admin@example.com", "$2a$10$abcdefghijklmnopqrstuv"); err != nil {
		t.Fatalf("Seed() unexpected error = %v", err)
	}

	must, err := svc.MustChangePassword(context.Background(), "admin@example.com")
	if err != nil {
		t.Fatalf("MustChangePassword() unexpected error = %v", err)
	}
	if !must {
		t.Error("MustChangePassword() = false, want true right after Seed()")
	}
}

func TestService_ChangePassword(t *testing.T) {
	t.Run("succeeds with the correct current password and a policy-compliant new one, clearing MustChangePassword", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)
		if err := svc.Register(context.Background(), "alice@example.com", "Temp0rary!", RoleMember, nil); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		if err := svc.ChangePassword(context.Background(), "alice@example.com", "Temp0rary!", "N3wPassw0rd!"); err != nil {
			t.Fatalf("ChangePassword() unexpected error = %v", err)
		}

		if !svc.Authenticate(context.Background(), "alice@example.com", "N3wPassw0rd!") {
			t.Error("Authenticate() with the new password = false, want true")
		}
		must, err := svc.MustChangePassword(context.Background(), "alice@example.com")
		if err != nil {
			t.Fatalf("MustChangePassword() unexpected error = %v", err)
		}
		if must {
			t.Error("MustChangePassword() = true after a successful ChangePassword(), want false")
		}
	})

	t.Run("rejects the wrong current password", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)
		if err := svc.Register(context.Background(), "alice@example.com", "Temp0rary!", RoleMember, nil); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		err := svc.ChangePassword(context.Background(), "alice@example.com", "wrong-password", "N3wPassw0rd!")
		if _, ok := err.(*errorutils.ValidationError); !ok {
			t.Errorf("ChangePassword() error = %v, want *errorutils.ValidationError", err)
		}
	})

	t.Run("rejects a new password that fails the security policy", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)
		if err := svc.Register(context.Background(), "alice@example.com", "Temp0rary!", RoleMember, nil); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		err := svc.ChangePassword(context.Background(), "alice@example.com", "Temp0rary!", "weak")
		if _, ok := err.(*errorutils.ValidationError); !ok {
			t.Errorf("ChangePassword() error = %v, want *errorutils.ValidationError", err)
		}
	})

	t.Run("rejects reusing the current password as the new one", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewUserService(repo)
		if err := svc.Register(context.Background(), "alice@example.com", "Temp0rary!", RoleMember, nil); err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		err := svc.ChangePassword(context.Background(), "alice@example.com", "Temp0rary!", "Temp0rary!")
		if _, ok := err.(*errorutils.ValidationError); !ok {
			t.Errorf("ChangePassword() error = %v, want *errorutils.ValidationError", err)
		}
	})
}

// TestService_RequirePasswordChange is a regression test for accounts
// created before MustChangePassword existed (no such field persisted,
// decodes as false) — an admin must be able to flag them so they're forced
// to change password on their next login, without touching their existing
// password hash.
func TestService_RequirePasswordChange(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	// Simulates a legacy record, written straight to the repository (as if
	// migrated from before this field existed).
	repo.byEmail["legacy@example.com"] = Entity{Email: "legacy@example.com", PasswordHash: "hash", Role: RoleMember}

	if err := svc.RequirePasswordChange(context.Background(), "legacy@example.com"); err != nil {
		t.Fatalf("RequirePasswordChange() unexpected error = %v", err)
	}

	must, err := svc.MustChangePassword(context.Background(), "legacy@example.com")
	if err != nil {
		t.Fatalf("MustChangePassword() unexpected error = %v", err)
	}
	if !must {
		t.Error("MustChangePassword() = false after RequirePasswordChange(), want true")
	}

	stored, err := repo.GetByEmail(context.Background(), "legacy@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() unexpected error = %v", err)
	}
	if stored.PasswordHash != "hash" {
		t.Errorf("PasswordHash = %q, want unchanged %q", stored.PasswordHash, "hash")
	}

	if err := svc.RequirePasswordChange(context.Background(), "ghost@example.com"); err == nil {
		t.Error("RequirePasswordChange() for an unknown email = nil error, want a not-found error")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "valid password", password: "Str0ng!Pass", wantErr: false},
		{name: "too short", password: "Sh0rt!", wantErr: true},
		{name: "missing uppercase", password: "n0uppercase!", wantErr: true},
		{name: "missing lowercase", password: "N0LOWERCASE!", wantErr: true},
		{name: "missing digit", password: "NoDigitsHere!", wantErr: true},
		{name: "missing special character", password: "NoSpecialChar1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateAccess(t *testing.T) {
	repo := newFakeRepository()
	svc := NewUserService(repo)

	if err := svc.Register(context.Background(), "alice@example.com", "pw", RoleMember, []string{"checkout"}); err != nil {
		t.Fatalf("Register() unexpected error = %v", err)
	}

	if err := svc.UpdateAccess(context.Background(), "alice@example.com", RoleMember, []string{"checkout", "billing"}); err != nil {
		t.Fatalf("UpdateAccess() unexpected error = %v", err)
	}

	stored, err := repo.GetByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() unexpected error = %v", err)
	}
	if len(stored.Projects) != 2 {
		t.Errorf("Projects = %v, want 2 entries", stored.Projects)
	}
	if stored.PasswordHash == "" {
		t.Error("UpdateAccess() must not clear the existing password hash")
	}
}
