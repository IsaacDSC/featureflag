package contenthub

import (
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	ID                 uuid.UUID          `json:"id" bson:"id"`
	Variable           string             `json:"key" bson:"key"`
	Value              string             `json:"value" bson:"value"`
	Description        string             `json:"description" bson:"description"`
	Active             bool               `json:"active" bson:"active"`
	CreatedAt          time.Time          `json:"created_at" bson:"created_at"`
	SessionsStrategies SessionsStrategies `json:"session_strategy" bson:"session_strategy"`
	BalancerStrategy   BalancerStrategy   `json:"balancer_strategy" bson:"balancer_strategy"`
}

func NewEntity(
	active bool,
	variable string,
	value string,
	description string,
	sessionStrategy SessionsStrategies,
	balancerStrategy BalancerStrategy,
) Entity {
	return Entity{
		ID:                 uuid.New(),
		Variable:           variable,
		Description:        description,
		Value:              value,
		Active:             active,
		SessionsStrategies: sessionStrategy,
		BalancerStrategy:   balancerStrategy,
		CreatedAt:          time.Now(),
	}
}
