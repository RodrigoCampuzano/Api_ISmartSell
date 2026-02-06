from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import List


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""
    
    # Database
    DATABASE_URL: str = "postgresql+asyncpg://user:password@localhost:5432/ismartsell"
    
    # JWT
    JWT_SECRET_KEY: str = "change-this-secret-key-in-production"
    JWT_ALGORITHM: str = "HS256"
    JWT_ACCESS_TOKEN_EXPIRE_MINUTES: int = 30
    
    # CORS
    CORS_ORIGINS: List[str] = ["http://localhost:3000", "http://localhost:8080"]
    
    # Reservation
    RESERVATION_TIMEOUT_MINUTES: int = 30
    
    # Payment
    PAYMENT_PROVIDER_API_KEY: str = ""
    PLATFORM_COMMISSION_RATE: float = 0.01
    
    # App
    PROJECT_NAME: str = "ISmartSell API"
    VERSION: str = "1.0.0"
    API_V1_PREFIX: str = "/api"
    
    model_config = SettingsConfigDict(
        env_file=".env",
        case_sensitive=True,
        extra="ignore"
    )


settings = Settings()
