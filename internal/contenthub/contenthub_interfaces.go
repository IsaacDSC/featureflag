package contenthub

import "context"

type Adapter interface {
	SaveContentHub(ctx context.Context, input Entity) error
	GetContentHub(ctx context.Context, key string) (Entity, error)
	GetAllContentHub(ctx context.Context) (map[string]Entity, error)
	DeleteContentHub(ctx context.Context, key string) error
}
