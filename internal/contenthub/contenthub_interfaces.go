package contenthub

import "context"

type Adapter interface {
	SaveContentHub(ctx context.Context, input Entity) error
	GetContentHub(ctx context.Context, key string) (Entity, error)
	GetAllContentHub(ctx context.Context) (map[string]Entity, error)
	DeleteContentHub(ctx context.Context, key string) error
}

// Auditor grava uma alteração de negócio na trilha de auditoria (email é lido
// do contexto pelo próprio implementador). Satisfeita por *audit.Service.
type Auditor interface {
	RecordChange(ctx context.Context, action, project, entityKey string) error
}
