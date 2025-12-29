package contenthub

import (
	"context"
	"fmt"
	"time"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/IsaacDSC/featureflag/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

const (
	collectionName = mongodb.CollectionName("contenthub")
	keyIndexModel  = mongodb.IndexModel("key")
)

func NewMongoDBContentHubRepository(database *mongo.Database) (*MongoDBRepository, error) {
	collection := database.Collection(collectionName.String())
	err := mongodb.CreateUniqueIndex(collection, keyIndexModel)
	if err != nil {
		return nil, fmt.Errorf("error on create index: %w", err)
	}

	return &MongoDBRepository{
		collection: collection,
		timeout:    10 * time.Second,
	}, nil
}

func (mr *MongoDBRepository) SaveContentHub(ctx context.Context, input Entity) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{keyIndexModel.String(): input.Variable}
	update := bson.M{
		"$set": bson.M{
			id:               input.ID,
			key:              input.Variable,
			value:            input.Value,
			description:      input.Description,
			active:           input.Active,
			createdAt:        input.CreatedAt,
			sessionStrategy:  input.SessionsStrategies,
			balancerStrategy: input.BalancerStrategy,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := mr.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (mr *MongoDBRepository) GetContentHub(ctx context.Context, key string) (Entity, error) {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{keyIndexModel.String(): key}
	var entity Entity

	err := mr.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Entity{}, errorutils.NewNotFoundError("contenthub")
		}
		return Entity{}, err
	}

	return entity, nil
}

func (mr *MongoDBRepository) GetAllContentHub(ctx context.Context) (map[string]Entity, error) {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	cursor, err := mr.collection.Find(ctx, bson.M{})
	if err != nil {
		return map[string]Entity{}, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]Entity)
	for cursor.Next(ctx) {
		var entity Entity
		if err := cursor.Decode(&entity); err != nil {
			return map[string]Entity{}, err
		}
		result[entity.Variable] = entity
	}

	if err := cursor.Err(); err != nil {
		return map[string]Entity{}, err
	}

	return result, nil
}

func (mr *MongoDBRepository) DeleteContentHub(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{"key": key}
	result, err := mr.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errorutils.NewNotFoundError("contenthub")
	}

	return nil
}
