from abc import abstractmethod
from typing import Optional
from src.core.domain.base_repository import BaseRepository
from src.modules.payments.domain.entities.payment import Payment


class PaymentRepository(BaseRepository[Payment]):
    """Payment repository interface (port)."""
    
    @abstractmethod
    async def get_by_order(self, order_id: int) -> Optional[Payment]:
        """Get payment by order ID."""
        pass
    
    @abstractmethod
    async def record_platform_revenue(self, payment_id: int, amount: float) -> bool:
        """Record platform revenue/commission."""
        pass
