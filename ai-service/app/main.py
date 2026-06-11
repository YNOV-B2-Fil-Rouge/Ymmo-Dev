"""Ymmo Data/AI microservice.

A small FastAPI app that reads the Ymmo database and exposes analytical
endpoints (market trends, popular listings, price estimation, strategic
zones). The Go API or the front-end call it over HTTP.
"""
from fastapi import FastAPI

from .routers import health

app = FastAPI(
    title="Ymmo Data/AI Service",
    version="1.0",
    description="Analytics & predictions for the Ymmo platform.",
)

app.include_router(health.router)
