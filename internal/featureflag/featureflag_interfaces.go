package featureflag

import "context"

type Adapter interface {
	SaveFF(ctx context.Context, project string, input Entity) error
	GetAllFF(ctx context.Context, project string) (map[string]Entity, error)
	GetFF(ctx context.Context, project, key string) (Entity, error)
	DeleteFF(ctx context.Context, project, key string) error
}
