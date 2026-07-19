package contenthub

import (
	"context"
	"fmt"

	"github.com/IsaacDSC/featureflag/pkg/ctxlog"
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

const (
	actionCreated = "created"
	actionUpdated = "updated"
	actionDeleted = "deleted"
)

type Service struct {
	repository Adapter
	auditor    Auditor
}

func NewContentHubService(repository Adapter, auditor Auditor) *Service {
	return &Service{repository: repository, auditor: auditor}
}

func (ch Service) CreateOrUpdate(ctx context.Context, contenthub Entity) error {
	data, err := ch.repository.GetContentHub(ctx, contenthub.Variable)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			if err := ch.repository.SaveContentHub(ctx, contenthub); err != nil {
				return err
			}
			ch.recordChange(ctx, actionCreated, contenthub.Variable)
			return nil
		default:
			return err
		}
	}

	data.Active = contenthub.Active

	if err := ch.repository.SaveContentHub(ctx, data); err != nil {
		return fmt.Errorf("error on save contenthub: %w", err)
	}

	ch.recordChange(ctx, actionUpdated, data.Variable)
	return nil
}

func (ch Service) RemoveContentHub(ctx context.Context, key string) error {
	if err := ch.repository.DeleteContentHub(ctx, key); err != nil {
		return err
	}

	ch.recordChange(ctx, actionDeleted, key)
	return nil
}

// recordChange grava a auditoria de forma best-effort: uma falha aqui não
// reverte a escrita principal nem falha a requisição do usuário, só é logada.
// project vem sempre vazio — content hub não tem essa dimensão (spec 001 §9).
func (ch Service) recordChange(ctx context.Context, action, entityKey string) {
	if err := ch.auditor.RecordChange(ctx, action, "", entityKey); err != nil {
		ctxlog.GetLogger(ctx).Error("failed to record audit trail", "error", err, "action", action, "entity_key", entityKey)
	}
}

func (ch Service) GetAllContentHub(ctx context.Context) (map[string]Entity, error) {
	return ch.repository.GetAllContentHub(ctx)
}

func (ch Service) GetContentHub(ctx context.Context, key string) (Entity, error) {
	contenthub, err := ch.repository.GetContentHub(ctx, key)
	if err != nil {
		return contenthub, err
	}

	return contenthub, nil
}
