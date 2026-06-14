"""Predictive endpoints (price estimate + sale-delay) powered by scikit-learn."""
import numpy as np
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from sklearn.linear_model import LinearRegression

from ..database import query_df

router = APIRouter(tags=["predictions"])

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
        x = comparables[["area"]].to_numpy()
        y = comparables["price"].to_numpy()
        model = LinearRegression().fit(x, y)
        predicted = float(model.predict(np.array([[req.area]]))[0])
        method, sample = "linear_regression", int(len(comparables))
    else:
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

    predicted = max(predicted, 0.0)
    return {
        "estimated_price": round(predicted, -2),
        "method": method,
        "sample_size": sample,
        "city": req.city,
        "category_id": req.category_id,
        "area": req.area,
    }


class DelayRequest(BaseModel):
    city: str
    category_id: int
    area: float = Field(gt=0, description="Surface in m²")
    price: float = Field(gt=0, description="Asking price in €")


# Days between a property going live (created_at) and the first offer recorded on it.
DELAY_SQL = """
    SELECT p.area  AS area,
           p.price AS price,
           DATEDIFF(COALESCE(s.offer_date, s.created_at), p.created_at) AS days
    FROM sale_files s
    JOIN properties p ON p.id = s.property_id
    WHERE p.area > 0
      AND DATEDIFF(COALESCE(s.offer_date, s.created_at), p.created_at) >= 0
"""


@router.post("/predict-delay")
def predict_delay(req: DelayRequest):
    """Predict how many days a property is likely to take to sell."""
    comparables = query_df(
        DELAY_SQL + " AND p.city = :city AND p.category_id = :cat",
        {"city": req.city, "cat": req.category_id},
    )

    if len(comparables) >= MIN_SAMPLES:
        x = comparables[["area", "price"]].to_numpy()
        y = comparables["days"].to_numpy()
        model = LinearRegression().fit(x, y)
        predicted = float(model.predict(np.array([[req.area, req.price]]))[0])
        method, sample = "linear_regression", int(len(comparables))
    else:
        wide = query_df(DELAY_SQL + " AND p.city = :city", {"city": req.city})
        if wide.empty:
            wide = query_df(DELAY_SQL + " AND p.category_id = :cat", {"cat": req.category_id})
        if wide.empty:
            wide = query_df(DELAY_SQL)
        if wide.empty:
            raise HTTPException(status_code=404, detail="not enough data to predict the sale delay")

        predicted = float(wide["days"].mean())
        method, sample = "historical_average", int(len(wide))

    predicted = max(predicted, 1.0)
    return {
        "estimated_days": int(round(predicted)),
        "method": method,
        "sample_size": sample,
        "city": req.city,
        "category_id": req.category_id,
    }
