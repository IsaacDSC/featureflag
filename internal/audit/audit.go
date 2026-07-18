package audit

import (
	"time"

	"github.com/google/uuid"
)

const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

type Entity struct {
	ID        uuid.UUID `json:"id" bson:"id"`
	Email     string    `json:"email" bson:"email"`
	Action    string    `json:"action" bson:"action"`
	Domain    string    `json:"domain" bson:"domain"`
	Project   string    `json:"project" bson:"project"`
	EntityKey string    `json:"entity_key" bson:"entity_key"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
