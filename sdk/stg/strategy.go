package stg

import (
	"fmt"
	"math"
)

type Strategy[T any] struct {
	WithStrategy bool         `json:"with_strategy"`
	SessionsID   map[string]T `json:"session_id"`
	Percent      float64      `json:"percent"`
	QtdCall      uint         `json:"qtd_call"`
}

/*
	Calculate -> This function Calculate

1% -> percent -> (100-1)/10 = 9.9
50% -> percent -> (100-50)/10 = 5
30% -> percent -> (100-30)/10 = 7
90% -> percent -> (100-90)/10 = 1
*/
const MaxCall = 10

// Calculate -> This function Calculate the value of the strategy
func (s *Strategy[T]) Calculate() uint {
	return uint(math.Ceil((float64(100) - s.Percent) / float64(MaxCall)))
}

func (s *Strategy[T]) SetQtdCall() {
	if s.QtdCall == MaxCall {
		s.QtdCall = 0
	} else {
		s.QtdCall += 1
	}
}

// Value -> This function return value by strategy
func (s *Strategy[T]) Value(sessionID string) (output T) {
	value, ok := s.SessionsID[sessionID]
	if !ok {
		return s.SessionsID["0"]
	}

	if ok && s.Percent <= 0 || s.Percent >= 100 {
		return value
	}

	return s.SessionsID[fmt.Sprintf("%d", s.Calculate())]
}

// Bool -> This function return active or inactive by strategy
func (s *Strategy[T]) StrategyBool(sessionID string) (output bool) {
	if len(s.SessionsID) == 0 {
		return s.Calculate() <= s.QtdCall
	}

	if _, ok := s.SessionsID[sessionID]; ok {
		return true
	}

	return false
}

func (s *Strategy[T]) Balancer() bool {
	isActive := s.Calculate() <= s.QtdCall
	s.QtdCall++

	return isActive
}
