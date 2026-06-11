"""Configuration loaded from environment variables (12-factor app).

Keeps DB credentials out of the code, so the same image runs in dev and prod.
"""
import os

from dotenv import load_dotenv

load_dotenv()


class Settings:
    DB_HOST = os.getenv("DB_HOST", "127.0.0.1")
    DB_PORT = os.getenv("DB_PORT", "3306")
    DB_USER = os.getenv("DB_USER", "ymmo_user")
    DB_PASSWORD = os.getenv("DB_PASSWORD", "ymmo_pass")
    DB_NAME = os.getenv("DB_NAME", "ymmo")

    @property
    def database_url(self) -> str:
        # PyMySQL driver, read-only analytical workload.
        return (
            f"mysql+pymysql://{self.DB_USER}:{self.DB_PASSWORD}"
            f"@{self.DB_HOST}:{self.DB_PORT}/{self.DB_NAME}"
        )


settings = Settings()
