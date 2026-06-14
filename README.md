# Ymmo Plateforme immobilière

## 📖 Table of Contents
1. [🤔 Project Presentation](#-project-presentation)
2. [🛠️ Prerequisites](#-prerequisites)
3. [🚀 Installation and Startup](#-installation-and-startup)
   1. [🌍 Global Startup (Full Orchestration)](#-global-startup-full-orchestration)
   2. [🧰 Independent Startup (Dev)](#-independent-startup-dev)
   3. [🧪 Demo Data (Seed)](#-demo-data-seed)
4. [🔐 Environment Variables (`.env`)](#-environment-variables-env)
5. [🐳 Docker Architecture and Security (Hardening)](#-docker-architecture-and-security-hardening)
   1. [📦 Services](#-services)
   2. [🛡️ Applied Hardening](#-applied-hardening)
   3. [💾 Docker Volumes](#-docker-volumes)
   4. [🗂️ Repository Architecture Diagram](#-repository-architecture-diagram)
6. [🔄 Git Workflow and Build](#-git-workflow-and-build)
7. [📎 Useful Commands](#-useful-commands)
8. [🌐 Authors](#-authors)
9. [🪢 Appendix](#-appendix)

## 🤔 Project Presentation
This repository hosts **Ymmo**, a centralized real-estate platform for a national agency network (head office in Aix-en-Provence + regional agencies). It lets clients browse and buy/sell properties, and gives staff dashboards and an AI module for strategic decisions (price trends, popular listings, target zones, price & sale-delay predictions).

It is a **monorepo** orchestrating four containerized services with a single `docker compose`:

- **Front-end** : `HTML · Tailwind CSS · JavaScript (vanilla, ES modules)`, served by **nginx**.
- **API** : **Go (Gin)** REST API, **stateless**, **JWT-secured**, **JSON over HTTP**, documented with **Swagger/OpenAPI**, persistence via **GORM**.
- **Data/AI** : **Python (FastAPI · pandas · scikit-learn)** microservice reading the database **read-only**.
- **Database** : **MariaDB** (normalized 3NF schema).

Strict separation of responsibilities:
- the Go API owns business logic and **all writes**;
- the Python service only **reads** the same database (dedicated read-only user) and runs the analysis;
- the AI service is **never exposed publicly** — the browser always goes through the Go API (`/api/v1/ai/*`), which relays to it on the internal Docker network;
- the front-end is static and talks to the API via `fetch`.

Internal Go architecture (layered, SOLID):
- `handlers/` : HTTP controllers (thin)
- `services/` : business logic
- `repositories/` : data access (GORM)
- `models/` : GORM entities
- `dto/` : validated request/response objects
- `middleware/` : JWT auth, RBAC
- `security/` : JWT creation/verification

## 🛠️ Prerequisites
[![Git](https://img.shields.io/badge/GIT-E44C30?style=for-the-badge&logo=git&logoColor=white)](https://git-scm.com/downloads)

[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/products/)

[![Docker Compose](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/products/)

## 🚀 Installation and Startup
Clone the repository, then move to its root:
```bash
git clone https://github.com/YNOV-B2-Fil-Rouge/Ymmo-Dev.git
cd Ymmo-Dev
```

### 🌍 Global Startup (Full Orchestration)
The whole stack is containerized. From the project root:
```bash
cp .env.example .env      # then fill in the secrets
docker compose up --build
```
This builds and starts the four services together:

| Service | Container | URL |
|---|---|---|
| Front-end | `ymmo-front` | http://localhost |
| API (Go) | `ymmo-api` | http://localhost:8080/api/v1 |
| Swagger UI | `ymmo-api` | http://localhost:8080/swagger/index.html |
| Health check | `ymmo-api` | http://localhost:8080/health |
| Database | `ymmo-db` | localhost:3306 |
| Data/AI (Python) | `ymmo-ai` | internal only (via `/api/v1/ai/*`) |

### 🧰 Independent Startup (Dev)
For front-end development you can serve the static pages with any local HTTP server (do **not** open them as `file://`, the API's CORS rejects it). The CORS allows `localhost:5173`:
```bash
cd frontend
python -m http.server 5173      # then http://localhost:5173
```
Make sure the API is running (`docker compose up -d db api ai`).

### 🧪 Demo Data (Seed)
On a **fresh volume**, the database initializes automatically in order:
1. `db/schema.sql` : tables + reference data (roles, departments, categories, access matrix)
2. `db/init-ai-user.sh` : the read-only AI database user
3. `db/seed.sql` : demo data (agencies, test accounts, properties)

To start over from scratch:
```bash
docker compose down -v        # removes the data volume
docker compose up --build     # re-provisions everything (schema + AI user + seed)
```

Test accounts (demo seed):

| Email | Password | Role |
|---|---|---|
| `buyer@ymmo.fr` | `buyer1234` | Buyer |
| `agent@ymmo.fr` | `agent1234` | Agent |
| `director@ymmo.fr` | `director1234` | Agency director |
| `hq@ymmo.fr` | `hq1234` | Head office |
| `it@ymmo.fr` | `it1234` | IT & Support |

A buyer can request to **become a seller** (Profile → "Devenir vendeur"); an agent approves it (Dashboard → "Demandes vendeur"), which promotes the account to seller.

## 🔐 Environment Variables (`.env`)
The `.env` file holds the secrets and is **not versioned** (`.env.example` is the template). Each environment uses its own secrets.

| Variable | Role |
|---|---|
| `MARIADB_ROOT_PASSWORD` | MariaDB root password |
| `DB_NAME` | database name (`ymmo`) |
| `DB_USER` / `DB_PASSWORD` | application DB user (read/write) used by the Go API |
| `AI_DB_USER` / `AI_DB_PASSWORD` | **read-only** DB user used by the Python service |
| `APP_ENV` | `development` or `production` |
| `JWT_SECRET` | JWT signing secret |
| `PUBLIC_API_URL` | public base URL used to build uploaded-photo links (default `http://localhost:8080`) |

```bash
cp .env.example .env
```

## 🐳 Docker Architecture and Security (Hardening)

### 📦 Services
Defined in `docker-compose.yml`:
1. `db` : `mariadb:11.4`, init scripts mounted in `/docker-entrypoint-initdb.d`, healthcheck.
2. `api` : multi-stage build (`golang:1.22-alpine` → `alpine:3.20`), runs `swag init` at build time, serves the REST API on `8080`, stores uploaded photos in a volume.
3. `ai` : `python:3.12-slim`, FastAPI/uvicorn, **only `expose: 8000`** (no host port → internal network only).
4. `frontend` : `nginx:1.27-alpine` serving the static files on `80`.

### 🛡️ Applied Hardening
- **Non-root** users in the API and AI images (`ymmo`), and nginx runs as its own user.
- **Read-only database user** for the AI service (`GRANT SELECT` only), created at DB init.
- **AI service not exposed** to the host: reachable only through the Go API proxy on the internal network.
- **No fallback secrets**: the Compose file and the AI config fail fast if a required variable is missing.
- **Healthchecks** on `db` and `api`, and `depends_on: condition: service_healthy` to enforce startup order.
- Swagger docs generated **at image build time** (`swag init`) so no manual step is needed.

### 💾 Docker Volumes
- `ymmo_db_data` : MariaDB data persistence.
- `ymmo_uploads` : uploaded property photos (mounted read-write on the API at `/app/uploads`, served at `/uploads`).

### 🗂️ Repository Architecture Diagram
```text
Ymmo-Dev/
|-- .env.example
|-- docker-compose.yml
|-- README.md
|-- db/
|   |-- schema.sql
|   |-- seed.sql
|   |-- init-ai-user.sh
|   `-- schema.mermaid
|-- backend/            # Go API  (see backend/README.md)
|   |-- Dockerfile
|   |-- cmd/api/
|   `-- internal/
|       |-- handlers/  services/  repositories/  models/
|       |-- dto/  middleware/  security/  router/  config/  database/
|-- ai-service/         # Python AI  (see ai-service/README.md)
|   |-- Dockerfile
|   `-- app/
|       |-- main.py  config.py  database.py
|       `-- routers/   # health, analytics, predictions
`-- frontend/           # Static front  (see frontend/README.md)
    |-- Dockerfile
    |-- nginx.conf
    |-- *.html
    `-- js/
```

## 🔄 Git Workflow and Build
There is no external CI/CD pipeline in this repository; quality is enforced through the workflow and the build:
- **Branch per feature**, documented commits, **Pull Requests** (no direct push to `main`).
- The Go image build runs `swag init` (regenerates the OpenAPI docs) then compiles a static binary; a broken build fails the image.
- `.env` is never committed (see `.gitignore`).

## 📎 Useful Commands
Start the full stack:
```bash
docker compose up --build
```
Stop services:
```bash
docker compose down
```
Stop and wipe the database volume (re-seeds on next start):
```bash
docker compose down -v
```
Rebuild a single service:
```bash
docker compose up -d --build api      # or ai / frontend
```
Follow a service's logs:
```bash
docker compose logs -f api
```

## 🌐 Authors
- [**LEFEBVRE Nino**](https://github.com/Ewoukouskous) — DEV (Go API, front-end, Data/AI)
- [**AMIARD Renaud**](https://github.com/RAmiard) — INFRA (network, servers, VMs)
- [**LASBENNES Lucas**](https://github.com/LucasAstley) — INFRA (network, servers, VMs)

## 🪢 Appendix
- 📦 [Back-end (Go API) — `backend/README.md`](backend/README.md)
- 🧠 [Data/AI service (Python) — `ai-service/README.md`](ai-service/README.md)
- 🖥️ [Front-end — `frontend/README.md`](frontend/README.md)
- 🗃️ [DrawSQL Schema Diagram](https://drawsql.app/teams/ynov-campus-toulouse/diagrams/test)
- 🗃️ Database schema diagram (Mermaid) — `db/schema.mermaid`
