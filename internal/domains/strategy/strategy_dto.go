package strategy

import (
	"errors"
)

type StrategyDto struct {
	SessionsID []string `json:"session_id,omitempty"`
	Percent    float64  `json:"percent,omitempty"`
}

func (s StrategyDto) ToDomain() (Strategy, error) {
	if s.Percent > float64(0) && len(s.SessionsID) > 0 {
		return Strategy{}, errors.New("invalid strategy, chosen strategy with session or strategy with percent")
	}

	if s.Percent > float64(0) || len(s.SessionsID) > 0 {
		sessions := map[string]bool{}

		for i := range s.SessionsID {
			sessions[s.SessionsID[i]] = true
		}

		return Strategy{
			SessionsID:   sessions,
			Percent:      s.Percent,
			QtdCall:      0,
			WithStrategy: true,
		}, nil
	}

	return Strategy{}, nil
}

func StrategyFromDomain(strategy Strategy) StrategyDto {
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
