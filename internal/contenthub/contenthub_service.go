package contenthub

import (
	"github.com/IsaacDSC/featureflag/pkg/errorutils"
)

type Service struct {
	repository Adapter
}

func NewContentHubService(repository Adapter) *Service {
	return &Service{repository: repository}
}

func (ch Service) CreateOrUpdate(contenthub Entity) error {
	database, err := ch.repository.GetContentHub(contenthub.Variable)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			if err := ch.repository.SaveContentHub(contenthub); err != nil {
				return err
			}
			return nil
		default:
			return err
		}
	}

	database.Active = contenthub.Active

	return ch.repository.SaveContentHub(database)
}

func (ch Service) RemoveContentHub(key string) error {
	return ch.repository.DeleteContentHub(key)
}

func (ch Service) GetAllContentHub() (map[string]Entity, error) {
	return ch.repository.GetAllContentHub()
}

func (ch Service) GetContentHub(key string) (Entity, error) {
	contenthub, err := ch.repository.GetContentHub(key)
	if err != nil {
		return contenthub, err
	}

	return contenthub, nil
}
