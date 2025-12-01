package contenthub

import (
	"encoding/json"
	"time"

	"github.com/IsaacDSC/featureflag/internal/contenthub"
	"github.com/IsaacDSC/featureflag/sdk/stg"
)

type Value []byte

type Content struct {
	Key              string                        `json:"key"`
	Description      string                        `json:"description"`
	CreatedAt        time.Time                     `json:"created_at"`
	Strategy         stg.Strategy[Value]           `json:"strategy"`
	SessionStrategy  contenthub.SessionsStrategies `json:"session_strategy"`
	BalancerStrategy contenthub.BalancerStrategy   `json:"balancer_strategy"`
}

func (ff Content) SessionValue(sessionID string) Value {
	value := ff.Strategy.Value(sessionID)
	ff.Strategy.SetQtdCall()
	return value
}

func (ff Content) Value() Value {
	value := ff.BalancerStrategy.Distribution()
	ff.Strategy.SetQtdCall()
	b, _ := json.Marshal(value)
	return b
}

func (v Value) Unmarshal(arg any) error {
	return json.Unmarshal(v, arg)
}
