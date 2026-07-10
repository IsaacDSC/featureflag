
## Feature Flag

> **Escopo por `project`.** Todas as rotas de feature flag são isoladas por um
> `project` no path (`/featureflag/{project}/...`). O `project` deve casar com
> `^[a-z0-9][a-z0-9-]{0,63}$` (letras minúsculas/dígitos, hífens, 1–64 chars).
> Flags de projetos diferentes são totalmente independentes.

### Rotas

| Método & Path | Descrição | Auth |
|---|---|---|
| `PATCH /featureflag/{project}` | Cria/atualiza uma flag | — |
| `GET /featureflag/{project}/all` | Lista todas as flags do project | — |
| `GET /featureflag/{project}/{key}` | Busca uma flag (header `session_id` opcional) | `SERVICE_CLIENT` |
| `GET /featureflag/{project}/sdk/{key}` | Leitura via SDK (retorna `{"status":"true"}`) | `SDK_CLIENT` |
| `DELETE /featureflag/{project}/{key}` | Remove uma flag | `SERVICE_CLIENT` |

A autorização é feita pelo header `Authorization` com o valor cru do token
(`SERVICE_CLIENT_AT` → `SERVICE_CLIENT`; `SDK_CLIENT_AT` → `SDK_CLIENT`). Nos
exemplos abaixo o project usado é `checkout`.

### Circuit Breaker Type Feature Flag

_Como criar uma flag do tipo circuit breaker, sem strategy, priorizando simplicidade_

```sh
curl -X PATCH http://localhost:3000/featureflag/checkout \
  -H "Content-Type: application/json" \
  -d '{"flag_name": "new_name_invalid", "description": "new_description", "active": true}'

```

### Example 2

_Como criar uma flag com 50%, ou seja, 50% das chamadas ficam ativas e 50% inativas_

```sh
curl -X PATCH http://localhost:3000/featureflag/checkout \
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

_Como criar uma flag com configuração de sessão, onde apenas quem tem a sessão recebe a flag como ativa_

```sh
curl -X PATCH http://localhost:3000/featureflag/checkout \
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

### Listar / buscar / remover

```sh
# Lista todas as flags do project
curl http://localhost:3000/featureflag/checkout/all

# Busca uma flag específica (requer SERVICE_CLIENT)
curl http://localhost:3000/featureflag/checkout/new_name_invalid \
  -H "Authorization: service" \
  -H "session_id: 34eec623-c9f2-494e-bf66-57a85139fd69"

# Leitura via SDK (requer SDK_CLIENT)
curl http://localhost:3000/featureflag/checkout/sdk/teste1 \
  -H "Authorization: sdk" \
  -H "session_id: 34eec623-c9f2-494e-bf66-57a85139fd69"

# Remove uma flag (requer SERVICE_CLIENT)
curl -X DELETE http://localhost:3000/featureflag/checkout/teste3 \
  -H "Authorization: service"
```

### Feature Flag Usage

O SDK é instanciado com o `host` e o `project` que ele deve observar. Ele busca
todas as flags do project na inicialização e mantém o cache atualizado via
polling periódico (`WithEventualConsistency` define o intervalo). Não há SSE:
mudanças são refletidas no próximo ciclo de refresh.

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
	ff := featureflag.NewFeatureFlagSDK("http://localhost:3000", "checkout")

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
