package featureflag

import (
	"time"

	"github.com/IsaacDSC/featureflag/internal/strategy"

	"github.com/google/uuid"
)

type Entity struct {
	ID         uuid.UUID         `json:"id" bson:"id"`
	FlagName   string            `json:"flag_name" bson:"flag_name"`
	Strategies strategy.Strategy `json:"strategy" bson:"strategy"`
	Active     bool              `json:"active" bson:"active"`
	CreatedAt  time.Time         `json:"created_at" bson:"created_at"`
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
