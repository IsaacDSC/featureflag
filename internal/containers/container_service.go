package containers

import (
	"github.com/IsaacDSC/featureflag/internal/domains/contenthub"
	"github.com/IsaacDSC/featureflag/internal/domains/featureflag"
)

type ServiceContainer struct {
	FeatureFlagService *featureflag.Service
	ContentHubService  *contenthub.Service
}

func NewServiceContainer(repositories RepositoryContainer) ServiceContainer {
	return ServiceContainer{
		FeatureFlagService: featureflag.NewFeatureflagService(repositories.FeatureFlagRepository),
		ContentHubService:  contenthub.NewContentHubService(repositories.ContentHubRepository),
	}
}
