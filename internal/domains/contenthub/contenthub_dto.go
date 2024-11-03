package contenthub

import (
	"github.com/IsaacDSC/featureflag/internal/domains/strategy"
	"github.com/google/uuid"
	"time"
)

type Dto struct {
	ID          uuid.UUID            `json:"id"`
	Variable    string               `json:"flag_name"`
	Value       string               `json:"value"`
	Description string               `json:"description"`
	Active      bool                 `json:"active"`
	CreatedAt   time.Time            `json:"created_at"`
	Strategy    strategy.StrategyDto `json:"strategy"`
}

func (c *Dto) ToDomain() (Entity, error) {
	strategy, err := c.Strategy.ToDomain()
	if err != nil {
		return Entity{}, err
	}

	return NewEntity(
		c.Active,
		c.Variable,
		c.Value,
		c.Description,
		strategy,
	), nil
}

func FromDomain(contenthub Entity) Dto {
	return Dto{
		ID:          contenthub.ID,
		Variable:    contenthub.Variable,
		Value:       contenthub.Value,
		Description: contenthub.Description,
		Active:      contenthub.Active,
		CreatedAt:   contenthub.CreatedAt,
		Strategy:    strategy.StrategyFromDomain(contenthub.Strategy),
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
