from sqlalchemy import Column, Integer, Enum, Numeric, String, ForeignKey, DateTime
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from src.core.infrastructure.database import Base
from src.modules.orders.domain.entities.order import OrderStatus, PaymentMethod
import enum


class OrderStatusEnum(enum.Enum):
    """SQL Alchemy enum for order status."""
    PENDING = "PENDING"
    RESERVED = "RESERVED"
    PAID = "PAID"
    READY = "READY"
    DELIVERED = "DELIVERED"
    CANCELLED = "CANCELLED"


class PaymentMethodEnum(enum.Enum):
    """SQLAlchemy enum for payment method."""
    ONLINE = "ONLINE"
    CASH = "CASH"
    NONE = "NONE"


class OrderModel(Base):
    """SQLAlchemy ORM model for orders table."""
    
    __tablename__ = "orders"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    buyer_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    store_id = Column(Integer, ForeignKey("stores.id"), nullable=False, index=True)
    status = Column(Enum(OrderStatusEnum), nullable=False, default=OrderStatusEnum.PENDING, index=True)
    total = Column(Numeric(12, 2), nullable=False)
    subtotal = Column(Numeric(12, 2), nullable=True)
    shipping = Column(Numeric(12, 2), nullable=False, default=0)
    payment_method = Column(Enum(PaymentMethodEnum), nullable=False, default=PaymentMethodEnum.NONE)
    qr_token = Column(String(128), unique=True, nullable=True, index=True)
    reserved_until = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    
    # Relationships
    items = relationship("OrderItemModel", back_populates="order", cascade="all, delete-orphan")


class OrderItemModel(Base):
    """SQLAlchemy ORM model for order_items table."""
    
    __tablename__ = "order_items"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    order_id = Column(Integer, ForeignKey("orders.id", ondelete="CASCADE"), nullable=False, index=True)
    product_id = Column(Integer, ForeignKey("products.id"), nullable=False)
    quantity = Column(Integer, nullable=False)
    unit_price = Column(Numeric(12, 2), nullable=False)
    total_price = Column(Numeric(12, 2), nullable=False)
    
    # Relationships
    order = relationship("OrderModel", back_populates="items")
