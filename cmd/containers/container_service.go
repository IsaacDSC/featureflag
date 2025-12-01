package containers

import (
	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
	"github.com/IsaacDSC/featureflag/pkg/pubsub"
)

type ServiceContainer struct {
	FeatureFlagService *featureflag.Service
	ContentHubService  *contenthub.Service
}

func NewServiceContainer(repositories RepositoryContainer, pub pubsub.Publisher) ServiceContainer {
	return ServiceContainer{
		FeatureFlagService: featureflag.NewFeatureflagService(repositories.FeatureFlagRepository, pub),
		ContentHubService:  contenthub.NewContentHubService(repositories.ContentHubRepository, pub),
	}
}
