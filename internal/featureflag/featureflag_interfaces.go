package featureflag

import "context"

type Adapter interface {
	SaveFF(ctx context.Context, project string, input Entity) error
	GetAllFF(ctx context.Context, project string) (map[string]Entity, error)
	GetFF(ctx context.Context, project, key string) (Entity, error)
	DeleteFF(ctx context.Context, project, key string) error
	ListProjects(ctx context.Context) ([]string, error)
}

// Auditor grava uma alteração de negócio na trilha de auditoria (email é lido
// do contexto pelo próprio implementador). Satisfeita por *audit.Service.
type Auditor interface {
	RecordChange(ctx context.Context, action, project, entityKey string) error
}
