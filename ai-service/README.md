# Ymmo — Data/AI service (Python)

Analytics microservice for the Ymmo platform.
Stack: **FastAPI · pandas · scikit-learn · SQLAlchemy (read-only)**.

It reads the same MariaDB as the Go API and exposes analytical endpoints.
It never writes to the database.

## Why a separate Python service?

The subject requires **data analysis in Python**, while the main API is in Go.
Instead of mixing the two, the analytical workload lives in its own service
(separation of concerns / SOLID), using Python's data ecosystem (pandas,
scikit-learn).

## Run locally

```bash
cd ai-service
python -m venv .venv
.venv\Scripts\activate          # Windows  (source .venv/bin/activate on macOS/Linux)
pip install -r requirements.txt
copy .env.example .env          # then adjust DB credentials
uvicorn app.main:app --reload --port 8000
```

Interactive docs (FastAPI auto-generates OpenAPI): http://localhost:8000/docs

```bash
curl http://localhost:8000/health
# {"status":"ok","service":"ymmo-ai","database":"up"}
```

## Layout

```
app/
  main.py        # FastAPI app
  config.py      # env-based settings
  database.py    # SQLAlchemy engine + pandas helper (read-only)
  routers/       # one file per group of endpoints
```
