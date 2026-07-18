package featureflag

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

func NewFeatureflagService(repository Adapter, auditor Auditor) *Service {
	return &Service{repository: repository, auditor: auditor}
}

func (ff Service) CreateOrUpdate(ctx context.Context, project string, featureflag Entity) error {
	flag, err := ff.repository.GetFF(ctx, project, featureflag.FlagName)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			if err := ff.repository.SaveFF(ctx, project, featureflag); err != nil {
				return err
			}
			ff.recordChange(ctx, actionCreated, project, featureflag.FlagName)
			return nil
		default:
			return err
		}
	}

	flag.Active = featureflag.Active
	flag.Strategies = featureflag.Strategies

	if err := ff.repository.SaveFF(ctx, project, flag); err != nil {
		return fmt.Errorf("error on save in repository: %w", err)
	}

	ff.recordChange(ctx, actionUpdated, project, flag.FlagName)
	return nil
}

func (ff Service) RemoveFeatureFlag(ctx context.Context, project, key string) error {
	if err := ff.repository.DeleteFF(ctx, project, key); err != nil {
		return err
	}

	ff.recordChange(ctx, actionDeleted, project, key)
	return nil
}

// recordChange grava a auditoria de forma best-effort: uma falha aqui não
// reverte a escrita principal nem falha a requisição do usuário, só é logada.
func (ff Service) recordChange(ctx context.Context, action, project, entityKey string) {
	if err := ff.auditor.RecordChange(ctx, action, project, entityKey); err != nil {
		ctxlog.GetLogger(ctx).Error("failed to record audit trail", "error", err, "action", action, "project", project, "entity_key", entityKey)
	}
}

func (ff Service) GetAllFeatureFlag(ctx context.Context, project string) (map[string]Entity, error) {
	return ff.repository.GetAllFF(ctx, project)
}

func (ff Service) ListProjects(ctx context.Context) ([]string, error) {
	return ff.repository.ListProjects(ctx)
}

func (ff Service) GetFeatureFlag(ctx context.Context, project, key string, sessionID string) (Entity, error) {
	featureflag, err := ff.repository.GetFF(ctx, project, key)
	if err != nil {
		return Entity{}, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(ctx, project, featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return Entity{}, err
		}
	}

	return featureflag, nil
}

func (ff Service) GetFeatureFlagBySDK(ctx context.Context, project, key string, sessionID string) (bool, error) {
	featureflag, err := ff.repository.GetFF(ctx, project, key)
	if err != nil {
		return false, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(ctx, project, featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return false, err
		}

		return featureflag.Strategies.IsActiveWithStrategy(sessionID), nil
	}

	return featureflag.Active, nil
}
