package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionName string

func (c CollectionName) String() string {
	return string(c)
}

type IndexModel string

func (c IndexModel) String() string {
	return string(c)
}

// CreateUniqueIndex creates a unique index on the collection with timeout of 2 seconds
func CreateUniqueIndex(collection *mongo.Collection, indexModel IndexModel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	idxModel := mongo.IndexModel{
		Keys:    bson.M{indexModel.String(): 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, idxModel)
	return err
}
