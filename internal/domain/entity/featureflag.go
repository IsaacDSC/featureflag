package entity

import (
	"github.com/google/uuid"
	"math"
	"time"
)

const MaxCall = 10

type Featureflag struct {
	ID         uuid.UUID `json:"id"`
	FlagName   string    `json:"flag_name"`
	Strategies Strategy  `json:"strategy"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ff Featureflag) SetStrategy(sessionID string) Featureflag {
	ff.Active = ff.Strategies.isActiveWithStrategy(sessionID)
	return ff
}

func (ff Featureflag) SetQtdCall() Featureflag {
	ff.Strategies.setQtdCall()
	return ff
}

func (ff Featureflag) IsUseStrategy() bool {
	return ff.Strategies.WithStrategy
}

type Strategy struct {
	WithStrategy bool            `json:"with_strategy"`
	SessionsID   map[string]bool `json:"session_id"`
	Percent      float64         `json:"percent"`
	QtdCall      uint            `json:"qtd_call"`
}

/*
	calculate -> This function calculate

1% -> percent -> (100-1)/10 = 9.9
50% -> percent -> (100-50)/10 = 5
30% -> percent -> (100-30)/10 = 7
90% -> percent -> (100-90)/10 = 1
*/
func (s *Strategy) calculate() uint {
	return uint(math.Ceil((float64(100) - s.Percent) / float64(10)))
}

func (s *Strategy) setQtdCall() {
	if s.QtdCall == MaxCall {
		s.QtdCall = 0
	} else {
		s.QtdCall += 1
	}
}

// isActiveWithStrategy -> This function return active or inactive by strategy
func (s *Strategy) isActiveWithStrategy(sessionID string) (output bool) {
	if len(s.SessionsID) == 0 {
		return s.calculate() <= s.QtdCall
	}

	if value, ok := s.SessionsID[sessionID]; ok {
		return value
	}

	return false
}
