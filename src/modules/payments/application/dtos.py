from pydantic import BaseModel
from typing import Optional
from src.modules.payments.domain.entities.payment import PaymentStatus


class PaymentRequest(BaseModel):
    """Request model for initiating payment."""
    provider: str = "stripe"


class PaymentResponse(BaseModel):
    """Response model for payment data."""
    id: int
    order_id: int
    amount: float
    provider: Optional[str]
    provider_fee: float
    platform_commission: float
    status: PaymentStatus
    
    class Config:
        from_attributes = True
