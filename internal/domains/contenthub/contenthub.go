package contenthub

import (
	"github.com/IsaacDSC/featureflag/internal/domains/strategy"
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	ID          uuid.UUID         `json:"id"`
	Variable    string            `json:"flag_name"`
	Value       string            `json:"value"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
	CreatedAt   time.Time         `json:"created_at"`
	Strategy    strategy.Strategy `json:"strategy"`
}

func NewEntity(
	active bool,
	variable string,
	value string,
	description string,
	strategy strategy.Strategy,
) Entity {
	return Entity{
		ID:          uuid.New(),
		Variable:    variable,
		Description: description,
		Value:       value,
		Active:      active,
		Strategy:    strategy,
		CreatedAt:   time.Now(),
	}
}
