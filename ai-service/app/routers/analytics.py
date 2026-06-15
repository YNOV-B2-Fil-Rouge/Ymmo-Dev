"""Market analytics endpoints (aggregations done in pandas)."""
from fastapi import APIRouter, Query

from ..database import query_df

router = APIRouter(tags=["analytics"])

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


def _normalize(s):
    """Min-max scale a pandas Series to [0, 1] (flat series -> 0)."""
    spread = s.max() - s.min()
    if spread == 0:
        return s * 0
    return (s - s.min()) / spread


@router.get("/zones")
def strategic_zones():
    """Rank cities by buying opportunity, to guide HQ's strategy.

    Score blends demand intensity (views per listing, higher = better) with
    price level (price per m², lower = better). Purely indicative.
    """
    df = query_df(
        f"""
        SELECT city, price, area, view_count
        FROM properties
        WHERE status IN {MARKET_STATUSES} AND area > 0
        """
    )
    if df.empty:
        return {"data": []}

    df["price_per_m2"] = df["price"] / df["area"]
    agg = (
        df.groupby("city")
        .agg(
            listings=("price", "count"),
            total_views=("view_count", "sum"),
            avg_price_per_m2=("price_per_m2", "mean"),
        )
        .reset_index()
    )
    agg["views_per_listing"] = (agg["total_views"] / agg["listings"]).round(1)

    agg["opportunity_score"] = (
        0.6 * _normalize(agg["views_per_listing"])
        + 0.4 * (1 - _normalize(agg["avg_price_per_m2"]))
    ).round(3)
    agg["avg_price_per_m2"] = agg["avg_price_per_m2"].round(0)

    agg = agg.sort_values("opportunity_score", ascending=False)
    return {"data": agg.to_dict(orient="records")}


@router.get("/dashboard/kpis")
def dashboard_kpis():
    """Global indicators for the head office dashboard."""
    props = query_df("SELECT status, price, area, view_count FROM properties")
    if props.empty:
        return {
            "total_properties": 0,
            "available": 0,
            "sold": 0,
            "avg_available_price": 0,
            "total_views": 0,
        }

    available = props[props["status"] == "AVAILABLE"]
    avg_price = float(available["price"].mean()) if not available.empty else 0.0

    return {
        "total_properties": int(len(props)),
        "available": int(len(available)),
        "sold": int((props["status"] == "SOLD").sum()),
        "avg_available_price": round(avg_price, -2),
        "total_views": int(props["view_count"].sum()),
    }
