# Ymmo Back-end API (Go)

## 📖 Table of Contents
1. [🤔 Module Overview](#-module-overview)
2. [🛠️ Prerequisites](#-prerequisites)
3. [🚀 Installation and Startup](#-installation-and-startup)
   1. [🌍 With Docker (recommended)](#-with-docker-recommended)
   2. [🧰 Local (without Docker)](#-local-without-docker)
4. [🔐 Environment Variables](#-environment-variables)
5. [🐳 Docker Architecture and Hardening](#-docker-architecture-and-hardening)
6. [📚 API Documentation (Swagger)](#-api-documentation-swagger)
7. [🗂️ Module Tree](#-module-tree)
8. [👥 Authors](#-authors)
9. [🪢 Appendix](#-appendix)

## 🤔 Module Overview
This is the **REST API** of the Ymmo platform, written in **Go (Gin)**.

The API is:
- **Stateless** : each request is autonomous.
- **JWT-secured** : authentication via signed JSON Web Tokens; role-based access control (RBAC).
- **JSON** : communicates with the front-end exclusively in JSON over HTTP.
- **Documented** : OpenAPI/Swagger annotations generate an interactive UI.

It owns the business logic and **all database writes** (the Python service only reads). It also serves uploaded property photos and proxies AI requests to the internal Python service.

Layered architecture (SOLID):
- `cmd/api/` : entry point (config → DB → router → HTTP server with graceful shutdown)
- `internal/handlers/` : HTTP controllers (thin)
- `internal/services/` : business logic
- `internal/repositories/` : data access (GORM)
- `internal/models/` : GORM entities (map `db/schema.sql`)
- `internal/dto/` : validated request/response objects
- `internal/middleware/` : JWT auth + RBAC
- `internal/security/` : JWT creation/verification
- `internal/router/` : Gin engine assembly (middleware + routes + dependency wiring)

## 🛠️ Prerequisites
[![Git](https://img.shields.io/badge/GIT-E44C30?style=for-the-badge&logo=git&logoColor=white)](https://git-scm.com/downloads)

[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/products/)

[![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=fff)](https://go.dev/dl/) *(local mode only)*

## 🚀 Installation and Startup

### 🌍 With Docker (recommended)
The API is part of the project's `docker-compose.yml`. From the **project root**:
```bash
docker compose up -d --build api
```
This builds the Go image (multi-stage, runs `swag init`), waits for the database to be healthy, then starts the API on `8080`.

### 🧰 Local (without Docker)
Requires Go 1.22+ and a running MariaDB seeded from `../db/schema.sql`.
```bash
cd backend
cp .env.example .env        # DB credentials + JWT_SECRET (DB_HOST=127.0.0.1)
go mod tidy
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/api/main.go -o docs
go run ./cmd/api            # http://localhost:8080
```

## 🔐 Environment Variables
| Variable | Role |
|---|---|
| `APP_ENV` | `development` / `production` |
| `APP_PORT` | HTTP port (default `8080`) |
| `JWT_SECRET` | JWT signing key (**required**) |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | MariaDB connection (read/write) |
| `AI_BASE_URL` | internal URL of the Python service (default `http://ai:8000`) |
| `UPLOAD_DIR` | on-disk folder for uploaded photos (default `/app/uploads`) |
| `PUBLIC_API_URL` | public base URL used to build photo URLs |

## 🐳 Docker Architecture and Hardening
- **Multi-stage build**: `golang:1.22-alpine` compiles a static binary, final image is `alpine:3.20` (small attack surface).
- **Swagger generated at build time** (`swag init`): the image is self-contained.
- Runs as a **non-root** user (`ymmo`); the upload directory is created and owned by that user so the mounted volume stays writable.
- **Healthcheck** hitting `/health`.
- Uploaded photos are served from the `ymmo_uploads` volume at `/uploads`.

## 📚 API Documentation (Swagger)
Interactive docs: **http://localhost:8080/swagger/index.html**

The spec is generated from the `// @...` annotations above each handler. With Docker it is automatic; in local mode regenerate after changing annotations:
```bash
swag init -g cmd/api/main.go -o docs
```

## 🗂️ Module Tree
```text
backend/
|-- Dockerfile
|-- go.mod / go.sum
|-- cmd/api/main.go
`-- internal/
    |-- config/        # env-based configuration
    |-- database/      # GORM / MariaDB connection
    |-- router/        # Gin engine + routes + wiring
    |-- middleware/    # JWT auth, RBAC
    |-- security/      # JWT
    |-- handlers/      # HTTP controllers (+ Swagger annotations)
    |-- services/      # business logic
    |-- repositories/  # data access
    |-- models/        # GORM entities
    `-- dto/           # validated payloads
```

## 👥 Authors
- **LEFEBVRE Nino** — DEV
- **AMIARD Renaud** — INFRA
- **LASBENNES Lucas** — INFRA

## 🪢 Appendix
- 🌍 [Global README](../README.md)
- 🧠 [Data/AI service](../ai-service/README.md)
- 🖥️ [Front-end](../frontend/README.md)
