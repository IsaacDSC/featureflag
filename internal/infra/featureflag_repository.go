package infra

import (
	"encoding/json"
	"github.com/IsaacDSC/featureflag/internal/domain/entity"
	"github.com/IsaacDSC/featureflag/internal/errorutils"
	"os"
)

type FeatureFlagRepository struct{}

func NewFeatureFlagRepository() *FeatureFlagRepository {
	return &FeatureFlagRepository{}
}

func (fr FeatureFlagRepository) SaveFF(input entity.Featureflag) error {
	featuresFlags, err := fr.GetAllFF()
	if err != nil {
		return err
	}

	featuresFlags[input.FlagName] = input
	b, err := json.Marshal(featuresFlags)
	if err != nil {
		return err
	}

	return os.WriteFile(FilePath, b, 0644)
}

func (fr FeatureFlagRepository) GetFF(key string) (entity.Featureflag, error) {
	b, err := os.ReadFile(FilePath)
	if err != nil {
		return entity.Featureflag{}, err
	}

	if len(b) == 0 {
		return entity.Featureflag{}, errorutils.NewNotFoundError("ff")
	}

	var ff map[string]entity.Featureflag
	if err := json.Unmarshal(b, &ff); err != nil {
		return entity.Featureflag{}, err
	}

	if output, ok := ff[key]; ok {
		return output, nil
	}

	return entity.Featureflag{}, errorutils.NewNotFoundError("ff")
}

func (fr FeatureFlagRepository) GetAllFF() (map[string]entity.Featureflag, error) {
	b, err := os.ReadFile(FilePath)
	if err != nil {
		return map[string]entity.Featureflag{}, err
	}

	if len(b) == 0 {
		return map[string]entity.Featureflag{}, nil
	}

	var ff map[string]entity.Featureflag
	if err := json.Unmarshal(b, &ff); err != nil {
		return map[string]entity.Featureflag{}, err
	}

	return ff, nil
}

func (fr FeatureFlagRepository) DeleteFF(key string) error {
	featuresflags, err := fr.GetAllFF()
	if err != nil {
		return err
	}

	delete(featuresflags, key)

	b, err := json.Marshal(featuresflags)
	if err != nil {
		return err
	}

	return os.WriteFile(FilePath, b, 0644)
}
