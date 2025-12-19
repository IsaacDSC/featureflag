### PATCH featureflag

```sh
  curl -X PATCH http://localhost:3000/featureflag -H "Content-Type: application/json" -d '{"flag_name": "new_name_invalid", "description": "new_description", "active": false}'
```

### GET featureflags

```sh
  curl -X GET http://localhost:3000/featureflags | jq
```

## Example

```sh
curl -i -X GET http://localhost:8080/health
```

### Create contenthub

```sh
curl -X PATCH http://localhost:3000/contenthub \
  -H "Content-Type: application/json" \
  --data-binary @example/contenthub/create_balancer_strategy.json
```
