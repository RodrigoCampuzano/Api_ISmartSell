from sqlalchemy import Column, Integer, String, Enum, DateTime
from sqlalchemy.sql import func
from src.core.infrastructure.database import Base
from src.modules.auth.domain.entities.user import UserRole
import enum


class UserRoleEnum(enum.Enum):
    """SQLAlchemy enum for user roles."""
    BUYER = "BUYER"
    SELLER = "SELLER"
    ADMIN = "ADMIN"


class UserModel(Base):
    """SQLAlchemy ORM model for users table."""
    
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    email = Column(String(255), unique=True, nullable=False, index=True)
    username = Column(String(100), unique=True, nullable=True, index=True)
    password_hash = Column(String(255), nullable=False)
    role = Column(Enum(UserRoleEnum), nullable=False, default=UserRoleEnum.BUYER)
    full_name = Column(String(200), nullable=True)
    phone = Column(String(30), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
