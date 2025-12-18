# Feature Flag

_Simples feature flag/ feature toagle_

![Green Teal Geometric Modern Computer Programmer Code Editor Quotes Instagram Post](https://github.com/IsaacDSC/featureflag/assets/56350331/eeb6227d-5a70-4a00-af21-e43368453c60)

## Startup service

_Set environment requirements using direnv_

```
export SECRET_KEY=""
export SERVICE_CLIENT_AT=""
export SDK_CLIENT_AT=""
```

_Start with docker project_

```sh
docker-compose up -d
```

## Install SDK

```sh
go get -u github.com/IsaacDSC/featureflag
```

## Authentication

_Get Auth SDK_

```
#### Get auth SDK client
POST http://localhost:3000/auth
Authorization: <token>
```

_Get Auth Service Client_

```
#### Get auth Service client
POST http://localhost:3000/auth
Authorization: <token>
```

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

## Contenthub

### Criando um contenthub

```sh
curl -X PATCH http://localhost:3000/contenthub \
  -H "Content-Type: application/json" \
  -d '{
  "key": "homepage_banner",
  "value": "enabled",
  "description": "Banner principal de la página de inicio con distribución ponderada",
  "active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "session_strategy": [
    {
      "session_id": "session-primers-123",
      "response": {
        "id": "promo-black-friday",
        "title": "Black Friday - Primers 50% OFF",
        "content": "¡Aprovecha nuestras ofertas de Black Friday!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Comprar Ahora",
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
        "content": "¡Aprovecha nuestras ofertas de Black Friday!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Comprar Ahora",
        "ctaUrl": "https://example.com/promo/black-friday",
        "backgroundColor": "#000000",
        "textColor": "#333333"
      }
    },
    {
      "session_id": "default",
      "response": {
        "id": "promo-black-friday",
        "title": "Fique atento as promoções",
        "content": "Não perca a chance de aproveitar nossas ofertas Black Friday!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Comprar Ahora",
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
        "content": "¡Aprovecha nuestras ofertas de Black Friday!",
        "imageUrl": "https://example.com/images/black-friday.jpg",
        "ctaText": "Comprar Ahora",
        "ctaUrl": "https://example.com/promo/black-friday",
        "backgroundColor": "#000000",
        "textColor": "#FFD700"
      }
    },
    {
      "weight": 10,
      "response": {
        "id": "new-users-welcome",
        "title": "Bienvenido Nuevo Usuario",
        "content": "Recibe 10% de descuento en tu primera compra",
        "imageUrl": "https://example.com/images/welcome.jpg",
        "ctaText": "Obtener Descuento",
        "ctaUrl": "https://example.com/welcome-offer",
        "backgroundColor": "#4A90E2",
        "textColor": "#FFFFFF"
      }
    },
    {
      "weight": 40,
      "response": {
        "id": "seasonal-collection",
        "title": "Nueva Colección de Temporada",
        "content": "Descubre las últimas tendencias de la temporada",
        "imageUrl": "https://example.com/images/seasonal.jpg",
        "ctaText": "Ver Colección",
        "ctaUrl": "https://example.com/collections/seasonal",
        "backgroundColor": "#F5F5F5",
        "textColor": "#333333"
      }
    }
  ]
}'
```

### Usage contenthub

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
