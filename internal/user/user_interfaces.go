package user

import "context"

type Adapter interface {
	Create(ctx context.Context, input Entity) error
	GetByEmail(ctx context.Context, email string) (Entity, error)
	Delete(ctx context.Context, email string) error
	ListAll(ctx context.Context) ([]Entity, error)
	// Update altera role/projects de um usuário já existente (nunca a senha).
	Update(ctx context.Context, input Entity) error
	// UpdatePassword grava um novo hash de senha e limpa MustChangePassword —
	// nunca toca role/projects.
	UpdatePassword(ctx context.Context, email, passwordHash string) error
	// RequirePasswordChange marca um usuário já existente para trocar a
	// senha no próximo login — usado por um admin para forçar essa troca em
	// contas criadas antes da feature existir, ou após suspeita de senha
	// comprometida. Nunca toca password_hash/role/projects.
	RequirePasswordChange(ctx context.Context, email string) error
}
