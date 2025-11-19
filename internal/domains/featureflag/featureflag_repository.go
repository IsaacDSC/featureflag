package featureflag

import (
	"encoding/json"
	"os"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type Repository struct{}

func NewFeatureFlagRepository() *Repository {
	return &Repository{}
}

func (fr Repository) SaveFF(input Entity) error {
	featuresFlags, err := fr.GetAllFF()
	if err != nil {
		return err
	}

	featuresFlags[input.FlagName] = input
	b, err := json.Marshal(featuresFlags)
	if err != nil {
		return err
	}

	return os.WriteFile(env.FilePath, b, 0644)
}

func (fr Repository) GetFF(key string) (Entity, error) {
	b, err := os.ReadFile(env.FilePath)
	if err != nil {
		return Entity{}, err
	}

	if len(b) == 0 {
		return Entity{}, errorutils.NewNotFoundError("ff")
	}

	var ff map[string]Entity
	if err := json.Unmarshal(b, &ff); err != nil {
		return Entity{}, err
	}

	if output, ok := ff[key]; ok {
		return output, nil
	}

	return Entity{}, errorutils.NewNotFoundError("ff")
}

func (fr Repository) GetAllFF() (map[string]Entity, error) {
	b, err := os.ReadFile(env.FilePath)
	if err != nil {
		return map[string]Entity{}, err
	}

	if len(b) == 0 {
		return map[string]Entity{}, nil
	}

	var ff map[string]Entity
	if err := json.Unmarshal(b, &ff); err != nil {
		return map[string]Entity{}, err
	}

	return ff, nil
}

func (fr Repository) DeleteFF(key string) error {
	featuresflags, err := fr.GetAllFF()
	if err != nil {
		return err
	}

	delete(featuresflags, key)

	b, err := json.Marshal(featuresflags)
	if err != nil {
		return err
	}

	return os.WriteFile(env.FilePath, b, 0644)
}
