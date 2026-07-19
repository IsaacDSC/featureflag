package user

import (
	"context"
	"fmt"
	"time"

	"github.com/IsaacDSC/featureflag/pkg/errorutils"
	"github.com/IsaacDSC/featureflag/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

const (
	collectionName = mongodb.CollectionName("users")
	emailIndexModel = mongodb.IndexModel("email")
)

func NewMongoDBUserRepository(database *mongo.Database) (*MongoDBRepository, error) {
	collection := database.Collection(collectionName.String())

	if err := mongodb.CreateUniqueIndex(collection, emailIndexModel); err != nil {
		return nil, fmt.Errorf("error on create index: %w", err)
	}

	return &MongoDBRepository{
		collection: collection,
		timeout:    10 * time.Second,
	}, nil
}

func (mr *MongoDBRepository) Create(ctx context.Context, input Entity) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	doc := bson.M{
		"_id":              input.ID,
		email:              input.Email,
		passwordHash:       input.PasswordHash,
		role:               input.Role,
		projects:           input.Projects,
		createdAt:          input.CreatedAt,
		mustChangePassword: input.MustChangePassword,
	}

	_, err := mr.collection.InsertOne(ctx, doc)
	return err
}

// Update altera só role/projects de um registro já existente — preserva
// password_hash, id e created_at do que já estava salvo.
func (mr *MongoDBRepository) Update(ctx context.Context, input Entity) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{emailIndexModel.String(): input.Email}
	update := bson.M{"$set": bson.M{role: input.Role, projects: input.Projects}}

	result, err := mr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errorutils.NewNotFoundError("user")
	}

	return nil
}

// UpdatePassword grava um novo hash de senha e limpa MustChangePassword —
// preserva role/projects (mesmo espírito de Update).
func (mr *MongoDBRepository) UpdatePassword(ctx context.Context, emailInput, passwordHashInput string) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{emailIndexModel.String(): emailInput}
	update := bson.M{"$set": bson.M{passwordHash: passwordHashInput, mustChangePassword: false}}

	result, err := mr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errorutils.NewNotFoundError("user")
	}

	return nil
}

// RequirePasswordChange marca um usuário já existente para trocar a senha
// no próximo login — preserva password_hash/role/projects.
func (mr *MongoDBRepository) RequirePasswordChange(ctx context.Context, emailInput string) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{emailIndexModel.String(): emailInput}
	update := bson.M{"$set": bson.M{mustChangePassword: true}}

	result, err := mr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errorutils.NewNotFoundError("user")
	}

	return nil
}

func (mr *MongoDBRepository) GetByEmail(ctx context.Context, emailInput string) (Entity, error) {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{emailIndexModel.String(): emailInput}
	var entity Entity

	err := mr.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Entity{}, errorutils.NewNotFoundError("user")
		}
		return Entity{}, err
	}

	return entity, nil
}

func (mr *MongoDBRepository) Delete(ctx context.Context, emailInput string) error {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	filter := bson.M{emailIndexModel.String(): emailInput}
	result, err := mr.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errorutils.NewNotFoundError("user")
	}

	return nil
}

func (mr *MongoDBRepository) ListAll(ctx context.Context) ([]Entity, error) {
	ctx, cancel := context.WithTimeout(ctx, mr.timeout)
	defer cancel()

	cursor, err := mr.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make([]Entity, 0)
	for cursor.Next(ctx) {
		var entity Entity
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		result = append(result, entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
