package strategy

import (
	"math"
)

const MaxCall = 10

type Strategy struct {
	WithStrategy bool            `json:"with_strategy"`
	SessionsID   map[string]bool `json:"session_id"`
	Percent      float64         `json:"percent"`
	QtdCall      uint            `json:"qtd_call"`
}

/*
	Calculate -> This function Calculate

1% -> percent -> (100-1)/10 = 9.9
50% -> percent -> (100-50)/10 = 5
30% -> percent -> (100-30)/10 = 7
90% -> percent -> (100-90)/10 = 1
*/
func (s *Strategy) Calculate() uint {
	return uint(math.Ceil((float64(100) - s.Percent) / float64(10)))
}

func (s *Strategy) SetQtdCall() {
	if s.QtdCall == MaxCall {
		s.QtdCall = 0
	} else {
		s.QtdCall += 1
	}
}

// IsActiveWithStrategy -> This function return active or inactive by strategy
func (s *Strategy) IsActiveWithStrategy(sessionID string) (output bool) {
	if len(s.SessionsID) == 0 {
		return s.Calculate() <= s.QtdCall
	}

	if value, ok := s.SessionsID[sessionID]; ok {
		return value
	}

	return false
}
