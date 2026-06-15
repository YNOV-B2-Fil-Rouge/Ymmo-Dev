from fastapi import APIRouter
from sqlalchemy import text

from ..database import engine

router = APIRouter(tags=["health"])


@router.get("/health")
def health():
    """Liveness + database reachability check."""
    db_status = "up"
    try:
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
    except Exception:
        db_status = "down"
    return {"status": "ok", "service": "ymmo-ai", "database": db_status}
