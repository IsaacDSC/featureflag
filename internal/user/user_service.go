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

// normalizeLegacyRole trata usuários persistidos antes do conceito de
// Role/Projects existir (campo "role" ausente/vazio no banco) como admin —
// é o equivalente exato do acesso irrestrito que todo usuário logado já
// tinha antes desta funcionalidade. Só se aplica a Role vazio (nunca a um
// valor inválido diferente disso), então não abre uma brecha de escalada de
// privilégio para dado novo: IsValidRole já bloqueia qualquer role vazio ou
// desconhecido nos caminhos de escrita (POST/PATCH /users).
func normalizeLegacyRole(e Entity) Entity {
	if e.Role == "" {
		e.Role = RoleAdmin
	}
	return e
}

func (s Service) getUser(ctx context.Context, emailInput string) (Entity, error) {
	u, err := s.repository.GetByEmail(ctx, emailInput)
	if err != nil {
		return Entity{}, err
	}
	return normalizeLegacyRole(u), nil
}

// Register cria um novo usuário a partir de uma senha em texto puro,
// recebida sobre uma sessão já autenticada de um admin (POST /users). O hash
// é calculado aqui, nunca pelo chamador.
func (s Service) Register(ctx context.Context, emailInput, password, roleInput string, projectsInput []string) error {
	if _, err := s.repository.GetByEmail(ctx, emailInput); err == nil {
		return errorutils.NewConflictError("user")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repository.Create(ctx, Entity{
		ID:                 uuid.New(),
		Email:              emailInput,
		PasswordHash:       string(hash),
		Role:               roleInput,
		Projects:           projectsInput,
		CreatedAt:          time.Now(),
		MustChangePassword: true,
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
	users, err := s.repository.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i] = normalizeLegacyRole(users[i])
	}
	return users, nil
}

func (s Service) GetByEmail(ctx context.Context, emailInput string) (Entity, error) {
	return s.getUser(ctx, emailInput)
}

// UpdateAccess altera o papel e os projects liberados para um usuário já
// existente — nunca a senha (troca de senha é fora de escopo, spec 003 §2).
func (s Service) UpdateAccess(ctx context.Context, emailInput, roleInput string, projectsInput []string) error {
	return s.repository.Update(ctx, Entity{
		Email:    emailInput,
		Role:     roleInput,
		Projects: projectsInput,
	})
}

// RequirePasswordChange força emailInput a trocar a senha no próximo login —
// usado por um admin para migrar contas criadas antes desta feature existir
// (sem MustChangePassword persistido) ou após suspeita de senha
// comprometida.
func (s Service) RequirePasswordChange(ctx context.Context, emailInput string) error {
	return s.repository.RequirePasswordChange(ctx, emailInput)
}

// IsAdmin é usado para gatilhar as rotas de gestão de usuários (POST/DELETE/
// GET/PATCH /users) a quem tem o papel admin. Um email desconhecido (JWT de
// um usuário já removido, por exemplo) não é admin.
func (s Service) IsAdmin(ctx context.Context, emailInput string) (bool, error) {
	u, err := s.getUser(ctx, emailInput)
	if err != nil {
		return false, nil
	}
	return u.Role == RoleAdmin, nil
}

// HasProjectAccess satisfaz featureflag.ProjectAccessChecker: admin acessa
// qualquer project; member só os que estiverem em Entity.Projects. Um email
// desconhecido (mesmo motivo de IsAdmin) não tem acesso a nada.
func (s Service) HasProjectAccess(ctx context.Context, emailInput, project string) (bool, error) {
	u, err := s.getUser(ctx, emailInput)
	if err != nil {
		return false, nil
	}
	return u.HasProjectAccess(project), nil
}

// Seed insere um hash já pronto (vindo de ADMIN_USERS), sem passar por
// GenerateFromPassword — usado só pelo bootstrap do primeiro usuário, nunca
// pela API de criação via POST /users. O usuário semeado é sempre admin: é
// ele quem vai criar/gerenciar os demais usuários e conceder acesso a
// projects.
func (s Service) Seed(ctx context.Context, emailInput, passwordHashInput string) error {
	return s.repository.Create(ctx, Entity{
		ID:                 uuid.New(),
		Email:              emailInput,
		PasswordHash:       passwordHashInput,
		Role:               RoleAdmin,
		CreatedAt:          time.Now(),
		MustChangePassword: true,
	})
}

// MustChangePassword reporta se emailInput precisa trocar a senha antes de
// acessar o resto do sistema (Entity.MustChangePassword). Um email
// desconhecido não bloqueia nada — mesmo critério "fail open" de IsAdmin e
// HasProjectAccess, já que o middleware que chama isso corre depois de
// RequireLogin/RequireServiceOrLogin já terem validado a sessão.
func (s Service) MustChangePassword(ctx context.Context, emailInput string) (bool, error) {
	u, err := s.getUser(ctx, emailInput)
	if err != nil {
		return false, nil
	}
	return u.MustChangePassword, nil
}

// ChangePassword troca a senha do próprio usuário autenticado — exige a
// senha atual (evita que uma sessão sequestrada troque a senha sem conhecer
// a original) e valida a nova senha contra ValidatePassword. Usado tanto
// pelo fluxo de primeiro login (MustChangePassword) quanto por uma troca de
// senha voluntária.
func (s Service) ChangePassword(ctx context.Context, emailInput, currentPassword, newPassword string) error {
	u, err := s.repository.GetByEmail(ctx, emailInput)
	if err != nil {
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(currentPassword)) != nil {
		return errorutils.NewValidationError("current password is incorrect")
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(newPassword)) == nil {
		return errorutils.NewValidationError("new password must be different from the current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repository.UpdatePassword(ctx, emailInput, string(hash))
}
