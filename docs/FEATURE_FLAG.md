
## Feature Flag

### Circuit Breaker Type Feature Flag

_How to create a circuit breaker type feature flag without strategy, prioritizing simplicity_

```sh
curl -X PATCH http://localhost:3000/featureflag -H "Content-Type: application/json" -d '{"flag_name": "new_name_invalid", "description": "new_description", "active": false}'
```

### Example 2

_How to create a feature flag with 50%, meaning 50% of calls will be active and 50% of calls will be inactive_

```sh
curl -X PATCH http://localhost:3000/featureflag \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "flag_name": "teste1",
    "active": true,
    "strategy": {
      "percent": 50
    }
  }'
```

### Example 3

_How to create a feature flag with session configurations, where only those with the session will receive the feature flag as enabled_

```sh
curl -X PATCH http://localhost:3000/featureflag \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "flag_name": "teste3",
    "active": true,
    "strategy": {
      "session_id": ["34eec623-c9f2-494e-bf66-57a85139fd69"]
    }
  }'
```

### Feature Flag Usage

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IsaacDSC/featureflag/sdk/featureflag"
)

func main() {

	ctx := context.Background()
	ff := featureflag.NewFeatureFlagSDK("http://localhost:3000")

	go func() {
		_, err := ff.Listenner(ctx)
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		isActive := ff.GetFeatureFlag("invalid_ff").WithDefault(true)
		fmt.Println("@@@", isActive)

		isActiveErr, err := ff.GetFeatureFlag("invalid_ff").Err()
		fmt.Println("@@@", isActiveErr, err)

		isActive2, err := ff.GetFeatureFlag("new_name").Err()
		fmt.Println("@@@", isActive2, err)

		isActive3 := ff.GetFeatureFlag("new_name1").Val()
		fmt.Println("@@@", isActive3)

		test1, err := ff.GetFeatureFlag("teste1").Err()
		fmt.Println("@@@ teste1: ", test1, err)

		test3, err := ff.GetFeatureFlag("teste3", "34eec623-c9f2-494e-bf66-57a85139fd69").Err()
		fmt.Println("@@@ teste3: ", test3, err)

		test4, err := ff.GetFeatureFlag("teste3", "not-found-session-id").Err()
		fmt.Println("@@@ teste4: ", test4, err)

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
```
