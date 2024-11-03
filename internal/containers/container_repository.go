package containers

import (
	"github.com/IsaacDSC/featureflag/internal/domains/contenthub"
	"github.com/IsaacDSC/featureflag/internal/domains/featureflag"
	"github.com/IsaacDSC/featureflag/internal/env"
)

type RepositoryContainer struct {
	FeatureFlagRepository featureflag.Adapter
	ContentHubRepository  contenthub.Adapter
}

func NewRepositoryContainer() RepositoryContainer {
	return RepositoryContainer{
		FeatureFlagRepository: featureflag.NewFeatureFlagRepository(),
		ContentHubRepository:  contenthub.NewContentHubRepository(env.FilePathContentHub),
	}
}
