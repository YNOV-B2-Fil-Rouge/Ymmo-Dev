"""Ymmo Data/AI microservice.

A small FastAPI app that reads the Ymmo database and exposes analytical
endpoints (market trends, popular listings, price estimation, strategic
zones). The Go API or the front-end call it over HTTP.
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .routers import analytics, health, predictions

app = FastAPI(
    title="Ymmo Data/AI Service",
    version="1.0",
    description="Analytics & predictions for the Ymmo platform.",
)

# Allow the front-end (served from another origin) to call the AI service.
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:5173"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(health.router)
app.include_router(analytics.router)
app.include_router(predictions.router)
