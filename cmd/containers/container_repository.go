package containers

import (
	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
	"go.mongodb.org/mongo-driver/mongo"
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

func NewRepositoryContainerMongodb(client *mongo.Client, mongodbName string) RepositoryContainer {
	database := client.Database(mongodbName)

	return RepositoryContainer{
		FeatureFlagRepository: featureflag.NewMongoDBFeatureFlagRepository(database, "featureflags"),
		ContentHubRepository:  contenthub.NewMongoDBContentHubRepository(database, "contenthub"),
	}
}
