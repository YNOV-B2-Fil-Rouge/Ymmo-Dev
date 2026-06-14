# Ymmo Data/AI Service (Python)

## 📖 Table of Contents
1. [🤔 Module Overview](#-module-overview)
2. [🛠️ Prerequisites](#-prerequisites)
3. [🚀 Installation and Startup](#-installation-and-startup)
   1. [🌍 With Docker (recommended)](#-with-docker-recommended)
   2. [🧰 Local (without Docker)](#-local-without-docker)
4. [🔐 Environment Variables](#-environment-variables)
5. [🧠 Endpoints](#-endpoints)
6. [🐳 Docker Architecture and Hardening](#-docker-architecture-and-hardening)
7. [🗂️ Module Tree](#-module-tree)
8. [👥 Authors](#-authors)
9. [🪢 Appendix](#-appendix)

## 🤔 Module Overview
This is the **data analysis / AI microservice** of Ymmo, written in **Python** with **FastAPI**, **pandas** and **scikit-learn**.

It reads the **same MariaDB** as the Go API, but **read-only** (dedicated `SELECT`-only user): the Go API owns all writes, this service never modifies data. The SQL only pulls raw rows; every aggregation, scoring and prediction is computed in pandas / scikit-learn, which is the Python data workflow required by the brief.

Why a separate service? The subject requires data analysis **in Python**, while the main API is in Go. Keeping the analytical workload in its own service (separation of concerns) lets us use Python's data ecosystem without mixing it into the API.

> 🔒 This service is **internal only**: it is not exposed on the host. The browser reaches it exclusively through the Go API proxy (`/api/v1/ai/*`).

## 🛠️ Prerequisites
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/products/)

[![Python](https://img.shields.io/badge/Python-3776AB?logo=python&logoColor=fff)](https://www.python.org/downloads/) 
*(local mode only)*

## 🚀 Installation and Startup

### 🌍 With Docker (recommended)
Part of the project's `docker-compose.yml`. From the **project root**:
```bash
docker compose up -d --build ai
```
The service starts once the database is healthy.

### 🧰 Local (without Docker)
Requires Python 3.11–3.13.
```bash
cd ai-service
python -m venv .venv
source .venv/bin/activate        # Windows: venv\Scripts\activate
pip install -r requirements.txt
cp .env.example .env             # DB credentials (read-only user)
uvicorn app.main:app --reload --port 8000
```
Auto-generated OpenAPI docs (local mode): http://localhost:8000/docs

## 🔐 Environment Variables
| Variable | Role |
|---|---|
| `DB_HOST` / `DB_PORT` | MariaDB host/port (`db` / `3306` in Docker) |
| `DB_USER` / `DB_PASSWORD` | **read-only** DB user (`AI_DB_USER` / `AI_DB_PASSWORD` from the global `.env`) |
| `DB_NAME` | database name (`ymmo`) |

The config fails fast if a required variable is missing (no fallback credentials).

## 🧠 Endpoints
Reached via the Go proxy at `/api/v1/ai/*`:

| Endpoint | Access | Description |
|---|---|---|
| `POST /estimate` | public | Price estimate (linear regression on comparables, avg €/m² fallback) |
| `POST /predict-delay` | public | Estimated days-to-sell (regression on area/price, historical-average fallback) |
| `GET /trends` | staff | Average €/m² by city/category (pandas) |
| `GET /popular` | staff | Most-viewed available properties |
| `GET /zones` | staff | Buying-opportunity score by city |
| `GET /dashboard/kpis` | staff | Global indicators (totals, available, sold, views) |
| `GET /health` | internal | Service + DB status |

## 🐳 Docker Architecture and Hardening
- Image `python:3.12-slim`, dependencies from `requirements.txt`.
- Runs as a **non-root** user (`ymmo`).
- **`expose: 8000` only** (no host port) — internal Docker network access only.
- Connects with a **read-only** database user; cannot write to the database.

## 🗂️ Module Tree
```text
ai-service/
|-- Dockerfile
|-- requirements.txt
`-- app/
    |-- main.py        # FastAPI app
    |-- config.py      # env-based settings (read-only DB user)
    |-- database.py    # SQLAlchemy engine + pandas helper (read-only)
    `-- routers/
        |-- health.py
        |-- analytics.py     # trends, popular, zones, dashboard/kpis
        `-- predictions.py   # estimate, predict-delay
```

## 👥 Authors
- **LEFEBVRE Nino** — DEV
- **AMIARD Renaud** — INFRA
- **LASBENNES Lucas** — INFRA

## 🪢 Appendix
- 🌍 [Global README](../README.md)
- 📦 [Back-end (Go API)](../backend/README.md)
- 🖥️ [Front-end](../frontend/README.md)
