package contenthub

import (
	"encoding/json"
	"time"

	"github.com/IsaacDSC/featureflag/internal/contenthub"
)

type Value []byte

type Content struct {
	Key              string                        `json:"key"`
	Description      string                        `json:"description"`
	CreatedAt        time.Time                     `json:"created_at"`
	SessionStrategy  contenthub.SessionsStrategies `json:"session_strategy"`
	BalancerStrategy contenthub.BalancerStrategy   `json:"balancer_strategy"`
}

func (ff Content) Value() Value {
	value := ff.BalancerStrategy.Distribution()
	b, _ := json.Marshal(value)
	return b
}
