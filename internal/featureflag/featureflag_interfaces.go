package featureflag

import "context"

type Adapter interface {
	SaveFF(ctx context.Context, input Entity) error
	GetAllFF(ctx context.Context) (map[string]Entity, error)
	GetFF(ctx context.Context, key string) (Entity, error)
	DeleteFF(ctx context.Context, key string) error
}
