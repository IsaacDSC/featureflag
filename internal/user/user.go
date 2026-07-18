package user

import (
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	ID           uuid.UUID `json:"id" bson:"id"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-" bson:"password_hash"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}
