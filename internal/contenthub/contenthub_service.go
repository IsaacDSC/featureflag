package contenthub

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

func NewContentHubService(repository Adapter, pub Publisher) *Service {
	return &Service{repository: repository, pub: pub}
}

func (ch Service) CreateOrUpdate(ctx context.Context, contenthub Entity) error {
	data, err := ch.repository.GetContentHub(ctx, contenthub.Variable)

	if err != nil {
		switch err.(type) {
		case *errorutils.NotFoundError:
			if err := ch.repository.SaveContentHub(ctx, contenthub); err != nil {
				return err
			}
			return nil
		default:
			return err
		}
	}

	data.Active = contenthub.Active

	if err := ch.repository.SaveContentHub(ctx, data); err != nil {
		return fmt.Errorf("error on save contenthub: %w", err)
	}

	if err := ch.pub.Publish(ctx, "contenthub", pubsub.NewPayload(data)); err != nil {
		return fmt.Errorf("error on publisher event writer contenthub: %w", err)
	}

	return nil
}

func (ch Service) RemoveContentHub(ctx context.Context, key string) error {
	return ch.repository.DeleteContentHub(ctx, key)
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
