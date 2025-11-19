package featureflag

import (
	"time"

	"github.com/IsaacDSC/featureflag/internal/strategy"

	"github.com/google/uuid"
)

type Entity struct {
	ID         uuid.UUID         `json:"id"`
	FlagName   string            `json:"flag_name"`
	Strategies strategy.Strategy `json:"strategy"`
	Active     bool              `json:"active"`
	CreatedAt  time.Time         `json:"created_at"`
}

func (ff Entity) SetStrategy(sessionID string) Entity {
	ff.Active = ff.Strategies.IsActiveWithStrategy(sessionID)
	return ff
}

func (ff Entity) SetQtdCall() Entity {
	ff.Strategies.SetQtdCall()
	return ff
}

func (ff Entity) IsUseStrategy() bool {
	return ff.Strategies.WithStrategy
}
