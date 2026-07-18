package featureflag

import (
	"context"
	"fmt"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type Service struct {
	repository Adapter
}

func NewFeatureflagService(repository Adapter) *Service {
	return &Service{repository: repository}
}

func (ff Service) CreateOrUpdate(ctx context.Context, project string, featureflag Entity) error {
	flag, err := ff.repository.GetFF(ctx, project, featureflag.FlagName)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			return ff.repository.SaveFF(ctx, project, featureflag)
		default:
			return err
		}
	}

	flag.Active = featureflag.Active
	flag.Strategies = featureflag.Strategies

	if err := ff.repository.SaveFF(ctx, project, flag); err != nil {
		return fmt.Errorf("error on save in repository: %w", err)
	}

	return nil
}

func (ff Service) RemoveFeatureFlag(ctx context.Context, project, key string) error {
	return ff.repository.DeleteFF(ctx, project, key)
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
