"""Database access for the analytics service.

We only READ from MariaDB and hand the data straight to pandas, which is where
the actual analysis happens. The Go API owns all writes; this service never
modifies data.
"""
import pandas as pd
from sqlalchemy import create_engine, text

from .config import settings

engine = create_engine(
    settings.database_url,
    pool_pre_ping=True,   # drop dead connections instead of failing a request
    pool_recycle=3600,
)


def query_df(sql: str, params: dict | None = None) -> pd.DataFrame:
    """Run a read-only SQL query and return the result as a pandas DataFrame."""
    with engine.connect() as conn:
        return pd.read_sql(text(sql), conn, params=params or {})
