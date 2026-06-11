"""Market analytics endpoints, powered by pandas.

The SQL only pulls raw rows; every aggregation/derivation is done in pandas to
showcase the Python data workflow required by the brief.
"""
from fastapi import APIRouter, Query

from ..database import query_df

router = APIRouter(tags=["analytics"])

# Property statuses that represent "real" market data (exclude drafts).
MARKET_STATUSES = ("AVAILABLE", "UNDER_OFFER", "SOLD")


@router.get("/trends")
def price_trends(city: str | None = Query(default=None, description="Filter on a city")):
    """Average price per m² by city (and category), computed with pandas.

    Helps advise users on where prices stand per sector.
    """
    df = query_df(
        """
        SELECT p.city, c.label AS category, p.price, p.area
        FROM properties p
        JOIN property_categories c ON c.id = p.category_id
        WHERE p.status IN ('AVAILABLE', 'UNDER_OFFER', 'SOLD')
          AND p.area > 0
        """
    )
    if df.empty:
        return {"data": []}

    if city:
        df = df[df["city"].str.lower() == city.lower()]
        if df.empty:
            return {"data": []}

    # Core computation: price per square meter.
    df["price_per_m2"] = df["price"] / df["area"]

    grouped = (
        df.groupby(["city", "category"])
        .agg(
            avg_price_per_m2=("price_per_m2", "mean"),
            avg_price=("price", "mean"),
            listings=("price", "count"),
        )
        .reset_index()
        .sort_values(["city", "avg_price_per_m2"], ascending=[True, False])
    )
    grouped["avg_price_per_m2"] = grouped["avg_price_per_m2"].round(0)
    grouped["avg_price"] = grouped["avg_price"].round(0)

    return {"data": grouped.to_dict(orient="records")}


@router.get("/popular")
def popular_properties(limit: int = Query(default=10, ge=1, le=50)):
    """Most-viewed available properties (popularity = view count)."""
    df = query_df(
        """
        SELECT id, reference, title, city, price, view_count
        FROM properties
        WHERE status = 'AVAILABLE'
        ORDER BY view_count DESC
        LIMIT :limit
        """,
        {"limit": limit},
    )
    return {"data": df.to_dict(orient="records")}
