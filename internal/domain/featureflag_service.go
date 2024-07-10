package domain

import (
	"github.com/IsaacDSC/featureflag/internal/domain/entity"
	"github.com/IsaacDSC/featureflag/internal/domain/interfaces"
	"github.com/IsaacDSC/featureflag/utils/errorutils"
)

type FeatureflagService struct {
	repository interfaces.FeatureFlagRepository
}

func NewFeatureflagService(repository interfaces.FeatureFlagRepository) *FeatureflagService {
	return &FeatureflagService{repository: repository}
}

func (ff FeatureflagService) CreateOrUpdate(featureflag entity.Featureflag) error {
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

func (ff FeatureflagService) RemoveFeatureFlag(key string) error {
	return ff.repository.DeleteFF(key)
}

func (ff FeatureflagService) GetAllFeatureFlag() (map[string]entity.Featureflag, error) {
	return ff.repository.GetAllFF()
}

func (ff FeatureflagService) GetFeatureFlag(key string, sessionID string) (entity.Featureflag, error) {
	featureflag, err := ff.repository.GetFF(key)
	if err != nil {
		return entity.Featureflag{}, err
	}

	if featureflag.IsUseStrategy() {
		if err := ff.repository.SaveFF(featureflag.SetStrategy(sessionID).SetQtdCall()); err != nil {
			return entity.Featureflag{}, err
		}
	}

	return featureflag, nil
}
