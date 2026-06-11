"""Predictive endpoints, powered by scikit-learn.

Price estimation is the predictive part required by the brief. We train a tiny
model on comparable properties; when there isn't enough data, we fall back to a
simple average price per m².
"""
import numpy as np
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from sklearn.linear_model import LinearRegression

from ..database import query_df

router = APIRouter(tags=["predictions"])

# Minimum comparable listings before we trust a trained model.
MIN_SAMPLES = 5
MARKET_FILTER = "status IN ('AVAILABLE', 'UNDER_OFFER', 'SOLD') AND area > 0"


class EstimateRequest(BaseModel):
    city: str
    category_id: int
    area: float = Field(gt=0, description="Surface in m²")


@router.post("/estimate")
def estimate_price(req: EstimateRequest):
    """Estimate a property's price from comparable listings.

    Strategy:
      - enough comparables (same city + category) -> linear regression on area;
      - otherwise -> average price per m² (city, then global) as a fallback.
    """
    comparables = query_df(
        f"""
        SELECT price, area FROM properties
        WHERE {MARKET_FILTER} AND city = :city AND category_id = :cat
        """,
        {"city": req.city, "cat": req.category_id},
    )

    if len(comparables) >= MIN_SAMPLES:
        # Train price = f(area) and predict for the requested area.
        x = comparables[["area"]].to_numpy()
        y = comparables["price"].to_numpy()
        model = LinearRegression().fit(x, y)
        predicted = float(model.predict(np.array([[req.area]]))[0])
        method, sample = "linear_regression", int(len(comparables))
    else:
        # Fallback: average price per m² in the city, then nationwide.
        wide = query_df(
            f"SELECT price, area FROM properties WHERE {MARKET_FILTER} AND city = :city",
            {"city": req.city},
        )
        if wide.empty:
            wide = query_df(f"SELECT price, area FROM properties WHERE {MARKET_FILTER}")
        if wide.empty:
            raise HTTPException(status_code=404, detail="not enough data to estimate")

        avg_price_per_m2 = float((wide["price"] / wide["area"]).mean())
        predicted = avg_price_per_m2 * req.area
        method, sample = "avg_price_per_m2", int(len(wide))

    predicted = max(predicted, 0.0)  # never return a negative estimate
    return {
        "estimated_price": round(predicted, -2),  # nearest 100
        "method": method,
        "sample_size": sample,
        "city": req.city,
        "category_id": req.category_id,
        "area": req.area,
    }
