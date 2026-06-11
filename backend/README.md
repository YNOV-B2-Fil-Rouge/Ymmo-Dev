# Ymmo — Back-end API (Go)

REST API for the Ymmo real-estate platform.
Stack: **Go (Gin) · GORM · MariaDB · JWT · Swagger** (added per module).

## Architecture (layered)

```
cmd/api            → entry point (main.go)
internal/
  config           → env-based configuration
  database         → GORM / MariaDB connection
  router           → Gin engine + route registration
  handlers         → HTTP controllers (thin)
  models           → GORM entities          (next module)
  repositories     → data access            (next module)
  services         → business logic         (next module)
  middleware       → JWT auth, RBAC, etc.    (next module)
  dto              → validated request/response objects (next module)
```

Each layer has one responsibility (SOLID). Handlers stay thin; logic
lives in services; data access is isolated in repositories.

## Run with Docker (recommended)

The whole stack (MariaDB + API) is containerized. From the **project root**:

```bash
docker compose up --build
```

This builds the Go image, starts MariaDB, loads `db/schema.sql` automatically
on first run, waits for the DB to be healthy, then starts the API.

```bash
curl http://localhost:8080/health   # {"database":"up",...}
docker compose down                 # stop
docker compose down -v              # stop + wipe the database volume
```

Defaults work out of the box; override credentials by copying
`../.env.example` to `../.env`.

## Run locally (without Docker)

### Prerequisites

- Go 1.22+
- MariaDB 10.6+ (database created from `../db/schema.sql`)

### Setup

```bash
# 1. Create the database (MariaDB client is `mariadb`; older installs may expose `mysql`)
mariadb -u root -p < ../db/schema.sql

# 2. Configure the app
cp .env.example .env
#   then edit .env (DB credentials + JWT_SECRET)

# 3. Install dependencies
go mod tidy

# 4. Run
go run ./cmd/api
```

## Health check

```bash
curl http://localhost:8080/health
# {"database":"up","service":"ymmo-api","status":"ok"}
```

## Conventions

- Branch per feature, documented commits, Pull Requests (no direct push to `main`).
- `.env` is never committed (see `.gitignore`).
