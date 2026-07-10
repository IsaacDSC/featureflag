package featureflag

import (
	"errors"
	"strings"
	"time"

	"github.com/IsaacDSC/featureflag/internal/strategy"
	"github.com/google/uuid"
)

type Dto struct {
	Project    string               `json:"project"`
	FlagName   string               `json:"flag_name"`
	Active     bool                 `json:"active"`
	Strategies strategy.StrategyDto `json:"strategy,omitempty"`
}

func ToDomain(project string, input Dto) (Entity, error) {
	if err := ValidateProject(project); err != nil {
		return Entity{}, err
	}

	if strings.TrimSpace(input.FlagName) == "" {
		return Entity{}, errors.New("flag name is required")
	}

	strategy, err := input.Strategies.ToDomain()
	if err != nil {
		return Entity{}, err
	}

	return Entity{
		ID:         uuid.New(),
		Project:    project,
		FlagName:   input.FlagName,
		Strategies: strategy,
		Active:     input.Active,
		CreatedAt:  time.Now(),
	}, nil
}

func DtoFromDomain(ff Entity) Dto {
	return Dto{
		Project:    ff.Project,
		FlagName:   ff.FlagName,
		Active:     ff.Active,
		Strategies: strategy.StrategyFromDomain(ff.Strategies),
	}
}
