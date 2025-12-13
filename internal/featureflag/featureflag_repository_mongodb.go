package featureflag

import (
	"context"
	"time"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewMongoDBFeatureFlagRepository(database *mongo.Database, collectionName string) *MongoDBRepository {
	return &MongoDBRepository{
		collection: database.Collection(collectionName),
		timeout:    10 * time.Second,
	}
}

func (mr *MongoDBRepository) SaveFF(input Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), mr.timeout)
	defer cancel()

	filter := bson.M{"flag_name": input.FlagName}
	update := bson.M{
		"$set": bson.M{
			"id":         input.ID,
			"flag_name":  input.FlagName,
			"strategy":   input.Strategies,
			"active":     input.Active,
			"created_at": input.CreatedAt,
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

	filter := bson.M{"flag_name": key}
	var entity Entity

	err := mr.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Entity{}, errorutils.NewNotFoundError("ff")
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

	filter := bson.M{"flag_name": key}
	result, err := mr.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errorutils.NewNotFoundError("ff")
	}

	return nil
}
