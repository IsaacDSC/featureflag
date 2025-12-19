
## Featureflag

### Featureflag do tipo disjuntor

_Como criar uma ff do tipo disjuntor sem strategy, priorizando a simplicidade_

```sh
curl -X PATCH http://localhost:3000/featureflag -H "Content-Type: application/json" -d '{"flag_name": "new_name_invalid", "description": "new_description", "active": false}'
```

### Example 2

_Como criar uma ff com 50% ou seja 50% das chamadas serão ativas e 50% das chamadas serão desativadas_

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

_Como criar uma ff com configurações utilizando sessions, onde somente quem estiver com a session receberá a feature
flag como ligada_

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

### Usage ff

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

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

```
