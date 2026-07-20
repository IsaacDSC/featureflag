package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/IsaacDSC/featureflag/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

const (
	collectionName      = mongodb.CollectionName("audit_log")
	createdAtIndexModel = mongodb.IndexModel("created_at")
)

func NewMongoDBAuditRepository(database *mongo.Database) (*MongoDBRepository, error) {
	collection := database.Collection(collectionName.String())

	// Índice não-único: uma trilha de auditoria não tem chave natural para
	// deduplicar, é sempre insert; o índice existe só para acelerar
	// consultas futuras por período.
	if err := mongodb.CreateIndex(collection, createdAtIndexModel); err != nil {
		return nil, fmt.Errorf("error on create index: %w", err)
	}

	return &MongoDBRepository{
		collection: collection,
		timeout:    10 * time.Second,
	}, nil
}

func (mr *MongoDBRepository) Record(ctx context.Context, entry Entity) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	doc := bson.M{
		id:        entry.ID,
		email:     entry.Email,
		action:    entry.Action,
		domain:    entry.Domain,
		project:   entry.Project,
		entityKey: entry.EntityKey,
		createdAt: entry.CreatedAt,
	}

	_, err := mr.collection.InsertOne(ctx, doc)
	return err
}
