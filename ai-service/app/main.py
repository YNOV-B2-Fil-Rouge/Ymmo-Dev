"""FastAPI Data/AI microservice: reads the Ymmo DB and exposes analytics endpoints."""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .routers import analytics, health, predictions

app = FastAPI(
    title="Ymmo Data/AI Service",
    version="1.0",
    description="Analytics & predictions for the Ymmo platform.",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:5173", "http://localhost:3001"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(health.router)
app.include_router(analytics.router)
app.include_router(predictions.router)
