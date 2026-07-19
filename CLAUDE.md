# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Feature-flag / feature-toggle service written in Go (stdlib `net/http`). It exposes two independent domains with different CAP trade-offs:

- **Feature Flag (AP)** — SDK caches flags in memory and refreshes them via periodic HTTP polling; the app always gets an answer even during network failures (eventual consistency).
- **Content Hub (CP)** — dynamic content served with strong consistency (always the freshest value).

The repo ships both the **server** (`cmd/`, `internal/`, `pkg/`) and the **client SDKs** (`sdk/`) that applications import via `go get github.com/IsaacDSC/featureflag`.

## Commands

```sh
go run ./cmd                 # run the server (listens on :3000)
go build -v ./...            # build everything
go test ./... --race         # full test suite (matches CI)
go test ./internal/featureflag -run TestName   # single package / single test
docker-compose up -d         # full stack: app + MongoDB + mongo-express (:8081)
make ci                      # run the same validations as CI locally (fmt, test, build, govulncheck)
```

CI (`.github/workflows/ci.yml`, Go 1.25) runs on pull requests and pushes to `main`/`master`/`develop`: `gofmt` check, `go test ./... --race`, `go build`, `govulncheck`. On push to `main` it additionally builds/pushes the Docker image, tagged `isaacdsc/featureflag:{short commit sha}`.

`loadtest/simple` and `loadtest/changed_status` are **separate Go modules** — `cd` into them before running/building; `go test ./...` at the root does not include them.

## Runtime prerequisites & gotchas

- **Required env vars** (validated by `cleanenv` in `internal/env/env.go`, `env-required`): `SECRET_KEY`, `SERVICE_CLIENT_AT`, `SDK_CLIENT_AT`. The server `log.Fatal`s at startup if any is missing. Note the README's `.env` sample omits the first two — `docker-compose.yml` supplies all of them.
- **`REPOSITORY_TYPE`** selects the persistence backend: `jsonfile` (default in code) writes to `featureflags.json` / `contenthub.json` in CWD; `mongodb` (used by docker-compose) requires `MONGODB_URI` and `MONGODB_NAME`. The check in `main.go` treats any value other than `"jsonfile"` as MongoDB.
- **Auth middleware is currently disabled at the top level.** `main.go` wraps every route with `middlewares.Logger` only (the `middlewares.Authorization` line is commented out). Per-route auth still applies where handlers wrap themselves — see below.

## Architecture

Request flow: `cmd/main.go` → `pkg/handlers.NewHandlers` merges route maps from each domain → routes registered on a single `http.ServeMux` using Go 1.22 method+path patterns (e.g. `PATCH /featureflag`, `GET /featureflag/sdk/{key}`).

Each domain (`internal/featureflag`, `internal/contenthub`) follows the same layering:

- `*_handle.go` — HTTP handlers; owns its route table and URL prefix. Feature-flag routes are individually wrapped with `middlewares.Authorization` + `middlewares.CheckPermission(...)`; content-hub routes have those wrappers commented out.
- `*_service.go` — business logic; depends on an `Adapter` (repository) interface and a `Publisher` interface, not concretions.
- `*_repository.go` (jsonfile) and `*_repository_mongodb.go` (MongoDB) — two implementations of the same `Adapter` interface (`*_interfaces.go`). Mock impls (`*_repository_mock.go`, `publisher_mock.go`) back the tests.
- `*_dto.go` / `*_fields.go` / `*.go` (Entity) — DTO ↔ domain conversion (`ToDomain` / `DtoFromDomain`) keeps wire format decoupled from the domain entity.

**Dependency wiring** lives in `cmd/containers/`:
- `container_repository.go` — `NewRepositoryContainer()` (jsonfile) vs `NewRepositoryContainerMongodb(client, name)`.
- `container_service.go` — builds `Service`s from the repository container + publisher.

**Updates via polling:**
- There is no server-push channel — flag changes are picked up by the SDK on its next poll. The service layer just persists to the repository on create/update; there is no pub/sub or SSE.
- The SDK (`sdk/featureflag/sdk.go`) on startup fetches all flags over HTTP into an in-memory map, then `Listenner` runs a periodic `refresh` ticker (interval set by `WithEventualConsistency`) that re-fetches all flags and diffs server vs memory (`filterChangedFlags` / `mergeFlags`) while preserving local strategy call counts (`QtdCall`). `Listenner` blocks until its context is cancelled, so it's typically run in a goroutine.

**Strategy evaluation** (percentage / session-based rollout) is shared logic used both server-side (`Entity.SetStrategy` / `IsActiveWithStrategy`) and in the SDK (`Flag.ValidateStrategy` / `Balancer` / `Increment`). Common strategy types live in `internal/strategy` and `sdk/stg`.

**Auth model** (`pkg/middlewares`): two static bearer tokens map to permissions — `SERVICE_CLIENT_AT` → `SERVICE_CLIENT` (write/admin), `SDK_CLIENT_AT` → `SDK_CLIENT` (SDK read via `/sdk/{key}` routes). `Authorization` reads the `Authorization` header and stores the resolved client name in the request context (`pkg/ctxutils`); `CheckPermission` gates the handler on it. `Authentication` is a separate cookie/JWT path (`pkg/authutils`) used by `/auth`.

**Cross-cutting `pkg/`:** `ctxlog` (slog logger carried in context, injected by `middlewares.Logger`), `ctxutils` (typed context get/set), `errorutils` (`NotFoundError` used for `errors.As`-style branching in services/handlers), `mongodb` (client/index helpers), `handlers` (route aggregation).

## HTTP reference

- Manual request collections: `featureflag_client.http`, `contenthub_client.http` (with `http-client.env.json`).
- Runnable examples: `example/featureflag/main.go`, `example/contenthub/main.go`.
- Domain docs: `docs/FEATURE_FLAG.md`, `docs/CONTENT_HUB.md`, `docs/ARCH.md`, `docs/USAGE_GUIDE.md`.
