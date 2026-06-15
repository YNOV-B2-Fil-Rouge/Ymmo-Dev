"""Read-only MariaDB access for the analytics service (results handed to pandas)."""
import pandas as pd
from sqlalchemy import create_engine, text

from .config import settings

engine = create_engine(
    settings.database_url,
    pool_pre_ping=True,
    pool_recycle=3600,
)


def query_df(sql: str, params: dict | None = None) -> pd.DataFrame:
    """Run a read-only SQL query and return the result as a pandas DataFrame."""
    with engine.connect() as conn:
        return pd.read_sql(text(sql), conn, params=params or {})
