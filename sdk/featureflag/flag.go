package featureflag

import "github.com/IsaacDSC/featureflag/sdk/stg"

type Flag struct {
	Active   bool
	FlagName string             `json:"flag_name"`
	Strategy stg.Strategy[bool] `json:"strategy"`
}

func (ff Flag) IsUseStrategy() bool {
	return ff.Strategy.WithStrategy
}

func (ff Flag) Increment() Flag {
	ff.Strategy.SetQtdCall()
	return ff
}

func (ff Flag) ValidateStrategy(sessionID string) Flag {
	ff.Active = ff.Strategy.StrategyBool(sessionID)
	return ff
}

func (ff Flag) Balancer() Flag {
	ff.Active = ff.Strategy.Balancer()
	return ff
}

func (ff Flag) Bool() bool {
	return ff.Active
}
