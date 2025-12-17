package featureflag

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
	collectionName     = mongodb.CollectionName("featureflags")
	flagNameIndexModel = mongodb.IndexModel("flag_name")
)

func NewMongoDBFeatureFlagRepository(database *mongo.Database) (*MongoDBRepository, error) {
	collection := database.Collection(collectionName.String())

	err := mongodb.CreateUniqueIndex(collection, flagNameIndexModel)
	if err != nil {
		return nil, fmt.Errorf("error on create index: %w", err)
	}

	return &MongoDBRepository{
		collection: collection,
		timeout:    10 * time.Second,
	}, nil
}

func (mr *MongoDBRepository) SaveFF(input Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), mr.timeout)
	defer cancel()

	filter := bson.M{flagNameIndexModel.String(): input.FlagName}
	update := bson.M{
		"$set": bson.M{
			id:         input.ID,
			flagName:   input.FlagName,
			strategies: input.Strategies,
			active:     input.Active,
			createdAt:  input.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := mr.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (mr *MongoDBRepository) GetFF(key string) (Entity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mr.timeout)
	defer cancel()

	filter := bson.M{flagNameIndexModel.String(): key}
	var entity Entity

	err := mr.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Entity{}, errorutils.NewNotFoundError("featureflag")
		}
		return Entity{}, err
	}

	return entity, nil
}

func (mr *MongoDBRepository) GetAllFF() (map[string]Entity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mr.timeout)
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
		result[entity.FlagName] = entity
	}

	if err := cursor.Err(); err != nil {
		return map[string]Entity{}, err
	}

	return result, nil
}

func (mr *MongoDBRepository) DeleteFF(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mr.timeout)
	defer cancel()

	filter := bson.M{flagNameIndexModel.String(): key}
	result, err := mr.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errorutils.NewNotFoundError("featureflag")
	}

	return nil
}
