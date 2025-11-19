package containers

import (
	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
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
