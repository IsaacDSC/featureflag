package contenthub

import (
	"encoding/json"
	"os"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type Repository struct {
	filePathContentHub string
}

func NewContentHubRepository(filePathContentHub string) *Repository {
	return &Repository{
		filePathContentHub: filePathContentHub,
	}
}

func (fr Repository) SaveContentHub(input Entity) error {
	featuresFlags, err := fr.GetAllContentHub()
	if err != nil {
		return err
	}

	featuresFlags[input.Variable] = input
	b, err := json.Marshal(featuresFlags)
	if err != nil {
		return err
	}

	return os.WriteFile(fr.filePathContentHub, b, 0644)
}

func (fr Repository) GetContentHub(key string) (Entity, error) {
	b, err := os.ReadFile(fr.filePathContentHub)
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

func (fr Repository) GetAllContentHub() (map[string]Entity, error) {
	b, err := os.ReadFile(fr.filePathContentHub)
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

func (fr Repository) DeleteContentHub(key string) error {
	featuresflags, err := fr.GetAllContentHub()
	if err != nil {
		return err
	}

	delete(featuresflags, key)

	b, err := json.Marshal(featuresflags)
	if err != nil {
		return err
	}

	return os.WriteFile(fr.filePathContentHub, b, 0644)
}
