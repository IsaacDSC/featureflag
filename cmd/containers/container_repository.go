package containers

import (
	"github.com/IsaacDSC/featureflag/internal/audit"
	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
	"github.com/IsaacDSC/featureflag/internal/user"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepositoryContainer struct {
	FeatureFlagRepository featureflag.Adapter
	ContentHubRepository  contenthub.Adapter
	UserRepository        user.Adapter
	AuditRepository       audit.Adapter
}

func NewRepositoryContainer() RepositoryContainer {
	return RepositoryContainer{
		FeatureFlagRepository: featureflag.NewFeatureFlagRepository(),
		ContentHubRepository:  contenthub.NewContentHubRepository(env.FilePathContentHub),
		UserRepository:        user.NewUserRepository(),
		AuditRepository:       audit.NewAuditRepository(),
	}
}

func NewRepositoryContainerMongodb(client *mongo.Client, mongodbName string) RepositoryContainer {
	database := client.Database(mongodbName)

	featureFlagRepository, err := featureflag.NewMongoDBFeatureFlagRepository(database)
	if err != nil {
		panic(err)
	}

	contentHubRepository, err := contenthub.NewMongoDBContentHubRepository(database)
	if err != nil {
		panic(err)
	}

	userRepository, err := user.NewMongoDBUserRepository(database)
	if err != nil {
		panic(err)
	}

	auditRepository, err := audit.NewMongoDBAuditRepository(database)
	if err != nil {
		panic(err)
	}

	return RepositoryContainer{
		FeatureFlagRepository: featureFlagRepository,
		ContentHubRepository:  contentHubRepository,
		UserRepository:        userRepository,
		AuditRepository:       auditRepository,
	}
}
