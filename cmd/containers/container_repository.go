package containers

import (
	"context"
	"log"
	"time"

	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/internal/featureflag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func NewRepositoryContainerMongodb() RepositoryContainer {
	environment := env.Get()

	// Create MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(environment.MongoDBURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	database := client.Database(environment.MongoDBName)

	// Initialize repositories with MongoDB collections
	return RepositoryContainer{
		FeatureFlagRepository: featureflag.NewMongoDBFeatureFlagRepository(database, "featureflags"),
		ContentHubRepository:  contenthub.NewMongoDBContentHubRepository(database, "contenthub"),
	}
}
