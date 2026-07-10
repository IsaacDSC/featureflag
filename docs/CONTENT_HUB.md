
## Content Hub

> **Modelo CP (strong consistency).** Diferente do Feature Flag, o Content Hub
> **não** é escopado por `project` e serve sempre o valor mais recente. As rotas
> de escrita/admin têm o middleware de `Authorization` **comentado no momento**,
> então nenhuma exige token hoje.

### Rotas

| Método & Path | Descrição | Chave lida de |
|---|---|---|
| `PATCH /contenthub` | Cria/atualiza um conteúdo | body (`key`) |
| `GET /contenthubs` | Lista todos os conteúdos | — |
| `GET /contenthub/{key}` | Busca um conteúdo | **query `?key=`** |
| `GET /contenthub/sdk/{key}` | Leitura via SDK | path `{key}` |
| `DELETE /contenthub/{key}` | Remove um conteúdo | **query `?key=`** |

> ⚠️ Atenção às inconsistências reais da API atual:
> - A lista fica em `/contenthub`**s** (plural), não `/contenthub`.
> - `GET /contenthub/{key}` e `DELETE /contenthub/{key}` ignoram o segmento do
>   path e leem a chave do query param `?key=`. O segmento precisa existir só
>   para casar a rota (ex.: `GET /contenthub/x?key=homepage_banner`). Apenas
>   `/contenthub/sdk/{key}` usa o valor do path.

### Creating a Content Hub

```sh
curl -X PATCH http://localhost:3000/contenthub \
  -H "Content-Type: application/json" \
  -d '{
  "key": "homepage_banner",
  "value": "enabled",
  "description": "Main homepage banner with weighted distribution",
  "active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "session_strategy": [
    {
      "session_id": "session-primers-123",
      "response": {
        "id": "promo-black-friday",
        "title": "Black Friday - Primers 50% OFF",
        "content": "Take advantage of our Black Friday deals!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Shop Now",
        "ctaUrl": "https://example.com/promo/black-friday",
        "backgroundColor": "#000000",
        "textColor": "#FFD700"
      }
    },
    {
      "session_id": "session-xyz-789",
      "response": {
        "id": "promo-black-friday",
        "title": "Black Friday - 10% OFF",
        "content": "Take advantage of our Black Friday deals!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Shop Now",
        "ctaUrl": "https://example.com/promo/black-friday",
        "backgroundColor": "#000000",
        "textColor": "#333333"
      }
    },
    {
      "session_id": "default",
      "response": {
        "id": "promo-black-friday",
        "title": "Stay tuned for promotions",
        "content": "Don't miss out on our Black Friday deals!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Shop Now",
        "ctaUrl": "https://example.com/promo/black-friday",
        "backgroundColor": "#000000",
        "textColor": "#FFFFFF"
      }
    }
  ],
  "balancer_strategy": [
    {
      "weight": 50,
      "response": {
        "id": "promo-black-friday",
        "title": "Black Friday - 50% OFF",
        "content": "Take advantage of our Black Friday deals!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Shop Now",
        "ctaUrl": "https://example.com/promo/black-friday",
        "backgroundColor": "#000000",
        "textColor": "#FFD700"
      }
    },
    {
      "weight": 10,
      "response": {
        "id": "new-users-welcome",
        "title": "Welcome New User",
        "content": "Get 10% off on your first purchase",
        "imageUrl": "https://example.com/images/welcome.jpg",
        "ctaText": "Get Discount",
        "ctaUrl": "https://example.com/welcome-offer",
        "backgroundColor": "#4A90E2",
        "textColor": "#FFFFFF"
      }
    },
    {
      "weight": 40,
      "response": {
        "id": "seasonal-collection",
        "title": "New Seasonal Collection",
        "content": "Discover the latest trends of the season",
        "imageUrl": "https://example.com/images/seasonal.jpg",
        "ctaText": "View Collection",
        "ctaUrl": "https://example.com/collections/seasonal",
        "backgroundColor": "#F5F5F5",
        "textColor": "#333333"
      }
    }
  ]
}'
```

### Listar / buscar / remover

```sh
# Lista todos os conteúdos (note o plural: /contenthubs)
curl http://localhost:3000/contenthubs

# Busca um conteúdo (a chave vem do query ?key=)
curl "http://localhost:3000/contenthub/homepage_banner?key=homepage_banner"

# Leitura via SDK (a chave vem do path)
curl http://localhost:3000/contenthub/sdk/homepage_banner

# Remove um conteúdo (a chave vem do query ?key=)
curl -X DELETE "http://localhost:3000/contenthub/homepage_banner?key=homepage_banner"
```

### Content Hub Usage

O SDK é instanciado apenas com o `host` (não há `project` no Content Hub). Ele
carrega todos os conteúdos na inicialização via `/contenthubs` e mantém o cache
atualizado via polling periódico (`WithEventualConsistency` define o intervalo).
Não há SSE.

```go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/IsaacDSC/featureflag/sdk/contenthub"
)

func main() {

	ctx := context.Background()
	contenthub := contenthub.NewContenthubSDK("http://localhost:3000")

	go func() {
		_, err := contenthub.Listenner(ctx)
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		content := contenthub.Content("invalid_ff").Val()
		fmt.Println("@@@", string(content))

		content1, err := contenthub.Content("invalid_ff").Err()
		fmt.Println("@@@", string(content1), err)

		content2, err := contenthub.Content("homepage_banner").Err()
		fmt.Println("@@@", string(content2), err)

		sessionsID := []string{"promo-black-friday", "new-users-welcome", "seasonal-collection"}
		random := rand.Intn(len(sessionsID))
		sessionID := sessionsID[random]

		content3, err := contenthub.Content("homepage_banner", sessionID).Err()
		fmt.Println("@@@", random, fmt.Sprintf("session_id:%s", sessionID), string(content3), err)

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
```
