# Feature Flag
*Simples feature flag/ feature toagle*

![Green Teal Geometric Modern Computer Programmer Code Editor Quotes Instagram Post](https://github.com/IsaacDSC/featureflag/assets/56350331/eeb6227d-5a70-4a00-af21-e43368453c60)

## Create service
```sh
docker run -p 3000:3000 isaacdsc/featureflag:v0.1
```

## SDK
```sh
go get -u github.com/IsaacDSC/featureflag
```


## Configuration
### Example 1
*Como criar uma ff*
```
###
PATCH http://localhost:3000/
Accept: application/json
Content-Type: application/json

{
  "flag_name": "teste1",
  "active": true
}
```

### Example 2
*Como criar uma ff com 50% ou seja 50% das chamadas serão ativas e 50% das chamadas serão desativadas*
```
###
PATCH http://localhost:3000/
Accept: application/json
Content-Type: application/json

{
  "flag_name": "teste1",
  "active": true,
  "strategy": {
    "percent": 50
  }
}
```

### Example 3
*Como criar uma ff com configurações utilizando sessions, onde somente quem estiver com a session receberá a feature flag como ligada*
```
###
PATCH http://localhost:3000/
Accept: application/json
Content-Type: application/json

{
  "flag_name": "teste3",
  "active": true,
  "strategy": {
     "session_id": ["34eec623-c9f2-494e-bf66-57a85139fd69"]
  }
}
```