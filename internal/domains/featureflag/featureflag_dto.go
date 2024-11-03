package featureflag

import (
	"github.com/IsaacDSC/featureflag/internal/domains/strategy"
	"github.com/google/uuid"
	"time"
)

type Dto struct {
	FlagName   string               `json:"flag_name"`
	Active     bool                 `json:"active"`
	Strategies strategy.StrategyDto `json:"strategy,omitempty"`
}

func ToDomain(input Dto) (Entity, error) {
	strategy, err := input.Strategies.ToDomain()
	if err != nil {
		return Entity{}, err
	}

	return Entity{
		ID:         uuid.New(),
		FlagName:   input.FlagName,
		Strategies: strategy,
		Active:     input.Active,
		CreatedAt:  time.Now(),
	}, nil
}

func DtoFromDomain(ff Entity) Dto {
	return Dto{
		FlagName:   ff.FlagName,
		Active:     ff.Active,
		Strategies: strategy.StrategyFromDomain(ff.Strategies),
	}
}
