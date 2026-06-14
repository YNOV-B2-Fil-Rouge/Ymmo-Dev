"""Configuration loaded from environment variables (12-factor app).

Keeps DB credentials out of the code, so the same image runs in dev and prod.
"""
import os

from dotenv import load_dotenv

load_dotenv()


def _require(key: str) -> str:
    """Return a required env variable, or fail fast if it is missing.

    No fallback credentials: the service refuses to start without explicit
    configuration.
    """
    value = os.getenv(key)
    if not value:
        raise RuntimeError(f"Missing required environment variable: {key}")
    return value


class Settings:
    DB_HOST = _require("DB_HOST")
    DB_PORT = os.getenv("DB_PORT", "3306")  # port is not a credential
    DB_USER = _require("DB_USER")
    DB_PASSWORD = _require("DB_PASSWORD")
    DB_NAME = _require("DB_NAME")

    @property
    def database_url(self) -> str:
        # PyMySQL driver, read-only analytical workload.
        return (
            f"mysql+pymysql://{self.DB_USER}:{self.DB_PASSWORD}"
            f"@{self.DB_HOST}:{self.DB_PORT}/{self.DB_NAME}"
        )


settings = Settings()
