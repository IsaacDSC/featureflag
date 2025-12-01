package contenthub

import (
	"errors"
)

type SessionStrategy struct {
	SessionID string `json:"session_id"`
	Response  any    `json:"response"`
}

type SessionsStrategies []SessionStrategy

var ErrInvalidSessionStrategy = errors.New("required default session strategy")

func (bs *SessionsStrategies) Validate() error {

	allReadyExistDefault := false
	for _, strategy := range *bs {
		if strategy.SessionID == "default" {
			allReadyExistDefault = true
		}
	}

	if allReadyExistDefault {
		return ErrInvalidSessionStrategy
	}

	return nil
}

func (bs SessionsStrategies) Val(sessionID string) any {
	var found bool
	var defaultResponse any
	var result any
	for _, strategy := range bs {
		if strategy.SessionID == sessionID {
			found = true
			result = strategy.Response
			break
		}
		if strategy.SessionID == "default" {
			defaultResponse = strategy.Response
		}
	}

	if !found {
		return defaultResponse
	}

	return result
}
