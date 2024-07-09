package dto

import (
	"errors"
	"github.com/IsaacDSC/featureflag/internal/domain/entity"
	"github.com/google/uuid"
	"time"
)

type FeatureflagDTO struct {
	FlagName   string      `json:"flag_name"`
	Active     bool        `json:"active"`
	Strategies StrategyDto `json:"strategy,omitempty"`
}

type StrategyDto struct {
	SessionsID []string `json:"session_id,omitempty"`
	Percent    float64  `json:"percent,omitempty"`
}

func FeatureFlagToDomain(input FeatureflagDTO) (entity.Featureflag, error) {
	strategy, err := StrategyToDomain(input.Strategies)
	if err != nil {
		return entity.Featureflag{}, err
	}

	return entity.Featureflag{
		ID:         uuid.New(),
		FlagName:   input.FlagName,
		Strategies: strategy,
		Active:     input.Active,
		CreatedAt:  time.Now(),
	}, nil
}

func StrategyToDomain(input StrategyDto) (entity.Strategy, error) {
	if input.Percent > float64(0) && len(input.SessionsID) > 0 {
		return entity.Strategy{}, errors.New("invalid strategy, chosen strategy with session or strategy with percent")
	}

	if input.Percent > float64(0) || len(input.SessionsID) > 0 {
		sessions := map[string]bool{}

		for i := range input.SessionsID {
			sessions[input.SessionsID[i]] = true
		}

		return entity.Strategy{
			SessionsID:   sessions,
			Percent:      input.Percent,
			QtdCall:      0,
			WithStrategy: true,
		}, nil
	}

	return entity.Strategy{}, nil
}

func StrategyFromDomain(strategy entity.Strategy) StrategyDto {
	sessions := make([]string, len(strategy.SessionsID))

	var counter int
	for key := range strategy.SessionsID {
		sessions[counter] = key
		counter++
	}

	return StrategyDto{
		SessionsID: sessions,
		Percent:    strategy.Percent,
	}
}

func FeatureFlagFromDomain(ff entity.Featureflag) FeatureflagDTO {
	return FeatureflagDTO{
		FlagName:   ff.FlagName,
		Active:     ff.Active,
		Strategies: StrategyFromDomain(ff.Strategies),
	}
}
