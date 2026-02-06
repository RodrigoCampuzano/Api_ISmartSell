from enum import Enum
from typing import Optional
from src.core.domain.base_entity import BaseEntity


class PaymentStatus(str, Enum):
    """Payment status states."""
    CREATED = "CREATED"
    COMPLETED = "COMPLETED"
    FAILED = "FAILED"
    REFUNDED = "REFUNDED"


class Payment(BaseEntity):
    """Payment domain entity."""
    
    def __init__(
        self,
        order_id: int,
        amount: float,
        status: PaymentStatus = PaymentStatus.CREATED,
        provider: Optional[str] = None,
        provider_fee: float = 0.0,
        platform_commission: float = 0.0,
        id: Optional[int] = None
    ):
        self.id = id
        self.order_id = order_id
        self.amount = amount
        self.provider = provider
        self.provider_fee = provider_fee
        self.platform_commission = platform_commission
        self.status = status
    
    def calculate_commission(self, commission_rate: float):
        """Calculate platform commission."""
        self.platform_commission = self.amount * commission_rate
