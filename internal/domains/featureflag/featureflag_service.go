package featureflag

import (
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type Service struct {
	repository Adapter
}

func NewFeatureflagService(repository Adapter) *Service {
	return &Service{repository: repository}
}

func (ff Service) CreateOrUpdate(featureflag Entity) error {
	database, err := ff.repository.GetFF(featureflag.FlagName)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			if err := ff.repository.SaveFF(featureflag); err != nil {
				return err
			}
			break
		default:
			return err
		}

		return nil
	}

	database.Active = featureflag.Active

	return ff.repository.SaveFF(database)
}

func (ff Service) RemoveFeatureFlag(key string) error {
	return ff.repository.DeleteFF(key)
}

func (ff Service) GetAllFeatureFlag() (map[string]Entity, error) {
	return ff.repository.GetAllFF()
}

func (ff Service) GetFeatureFlag(key string, sessionID string) (Entity, error) {
	featureflag, err := ff.repository.GetFF(key)
	if err != nil {
		return Entity{}, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return Entity{}, err
		}
	}

	return featureflag, nil
}

func (ff Service) GetFeatureFlagBySDK(key string, sessionID string) (bool, error) {
	featureflag, err := ff.repository.GetFF(key)
	if err != nil {
		return false, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return false, err
		}

		return featureflag.Strategies.IsActiveWithStrategy(sessionID), nil
	}

	return featureflag.Active, nil
}
