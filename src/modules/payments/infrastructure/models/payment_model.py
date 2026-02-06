from sqlalchemy import Column, Integer, String, Numeric, Enum, ForeignKey, DateTime
from sqlalchemy.sql import func
from src.core.infrastructure.database import Base
from src.modules.payments.domain.entities.payment import PaymentStatus
import enum


class PaymentStatusEnum(enum.Enum):
    """SQLAlchemy enum for payment status."""
    CREATED = "CREATED"
    COMPLETED = "COMPLETED"
    FAILED = "FAILED"
    REFUNDED = "REFUNDED"


class PaymentModel(Base):
    """SQLAlchemy ORM model for payments table."""
    
    __tablename__ = "payments"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    order_id = Column(Integer, ForeignKey("orders.id"), nullable=False, index=True)
    amount = Column(Numeric(12, 2), nullable=False)
    provider = Column(String(100), nullable=True)
    provider_fee = Column(Numeric(12, 2), nullable=False, default=0)
    platform_commission = Column(Numeric(12, 2), nullable=False, default=0)
    status = Column(Enum(PaymentStatusEnum), nullable=False, default=PaymentStatusEnum.CREATED)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)


class PlatformRevenueModel(Base):
    """SQLAlchemy ORM model for platform_revenues table."""
    
    __tablename__ = "platform_revenues"
    
    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    payment_id = Column(Integer, ForeignKey("payments.id"), nullable=False, index=True)
    amount = Column(Numeric(12, 2), nullable=False)
    collected_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
