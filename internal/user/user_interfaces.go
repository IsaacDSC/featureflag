package user

import "context"

type Adapter interface {
	Create(ctx context.Context, input Entity) error
	GetByEmail(ctx context.Context, email string) (Entity, error)
	Delete(ctx context.Context, email string) error
	ListAll(ctx context.Context) ([]Entity, error)
}
