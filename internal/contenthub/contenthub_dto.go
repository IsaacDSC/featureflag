package contenthub

import (
	"time"

	"github.com/google/uuid"
)

type Dto struct {
	ID                 uuid.UUID          `json:"id"`
	Variable           string             `json:"key"`
	Value              string             `json:"value"`
	Description        string             `json:"description"`
	Active             bool               `json:"active"`
	CreatedAt          time.Time          `json:"created_at"`
	SessionsStrategies SessionsStrategies `json:"session_strategy"`
	BalancerStrategy   BalancerStrategy   `json:"balancer_strategy"`
}

func (c *Dto) ToDomain() (Entity, error) {
	// sessionStrategy, err := c.SessionStrategy.ToDomain()
	// if err != nil {
	// 	return Entity{}, err
	// }

	if err := c.BalancerStrategy.Validate(); err != nil {
		return Entity{}, err
	}

	return NewEntity(
		c.Active,
		c.Variable,
		c.Value,
		c.Description,
		c.SessionsStrategies,
		c.BalancerStrategy,
	), nil
}

func FromDomain(contenthub Entity) Dto {
	return Dto{
		ID:                 contenthub.ID,
		Variable:           contenthub.Variable,
		Value:              contenthub.Value,
		Description:        contenthub.Description,
		Active:             contenthub.Active,
		CreatedAt:          contenthub.CreatedAt,
		SessionsStrategies: contenthub.SessionsStrategies,
		BalancerStrategy:   contenthub.BalancerStrategy,
	}
}

func ManyFromDomain(contenthub map[string]Entity) []Dto {
	output := make([]Dto, len(contenthub))
	counter := 0
	for _, content := range contenthub {
		output[counter] = FromDomain(content)
		counter++
	}
	return output
}
