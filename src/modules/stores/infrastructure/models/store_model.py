from sqlalchemy import Column, Integer, String, Text, Numeric, ForeignKey, DateTime
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from src.core.infrastructure.database import Base


class StoreModel(Base):
    """SQLAlchemy ORM model for stores table."""
    
    __tablename__ = "stores"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    seller_id = Column(Integer, ForeignKey("users.id", ondelete="CASCADE"), nullable=False, index=True)
    name = Column(String(255), nullable=False)
    slug = Column(String(255), unique=True, nullable=True, index=True)
    description = Column(Text, nullable=True)
    address = Column(String(500), nullable=True)
    lat = Column(Numeric(10, 8), nullable=True)
    lng = Column(Numeric(11, 8), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    
    # Relationships
    delivery_points = relationship("DeliveryPointModel", back_populates="store", cascade="all, delete-orphan")


class DeliveryPointModel(Base):
    """SQLAlchemy ORM model for store_delivery_points table."""
    
    __tablename__ = "store_delivery_points"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    store_id = Column(Integer, ForeignKey("stores.id", ondelete="CASCADE"), nullable=False, index=True)
    name = Column(String(255), nullable=False)
    address = Column(String(500), nullable=True)
    lat = Column(Numeric(10, 8), nullable=True)
    lng = Column(Numeric(11, 8), nullable=True)
    
    # Relationships
    store = relationship("StoreModel", back_populates="delivery_points")
