package featureflag

import (
	"context"
	"fmt"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/IsaacDSC/featureflag/pkg/pubsub"
)

type Publisher interface {
	Publish(ctx context.Context, channel string, msg pubsub.Payload) error
}

type Service struct {
	repository Adapter
	pub        Publisher
}

func NewFeatureflagService(repository Adapter, pub Publisher) *Service {
	return &Service{repository: repository, pub: pub}
}

func (ff Service) CreateOrUpdate(ctx context.Context, featureflag Entity) error {
	flag, err := ff.repository.GetFF(ctx, featureflag.FlagName)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			if err := ff.repository.SaveFF(ctx, featureflag); err != nil {
				return err
			}
		default:
			return err
		}

		return nil
	}

	flag.Active = featureflag.Active

	if err := ff.repository.SaveFF(ctx, flag); err != nil {
		return fmt.Errorf("error on save in repository: %w", err)
	}

	if err := ff.pub.Publish(ctx, "featureflag", pubsub.NewPayload(flag)); err != nil {
		return fmt.Errorf("error on publisher event writer feature flag: %w", err)
	}

	return nil
}

func (ff Service) RemoveFeatureFlag(ctx context.Context, key string) error {
	return ff.repository.DeleteFF(ctx, key)
}

func (ff Service) GetAllFeatureFlag(ctx context.Context) (map[string]Entity, error) {
	return ff.repository.GetAllFF(ctx)
}

func (ff Service) GetFeatureFlag(ctx context.Context, key string, sessionID string) (Entity, error) {
	featureflag, err := ff.repository.GetFF(ctx, key)
	if err != nil {
		return Entity{}, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(ctx, featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return Entity{}, err
		}
	}

	return featureflag, nil
}

func (ff Service) GetFeatureFlagBySDK(ctx context.Context, key string, sessionID string) (bool, error) {
	featureflag, err := ff.repository.GetFF(ctx, key)
	if err != nil {
		return false, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(ctx, featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return false, err
		}

		return featureflag.Strategies.IsActiveWithStrategy(sessionID), nil
	}

	return featureflag.Active, nil
}
