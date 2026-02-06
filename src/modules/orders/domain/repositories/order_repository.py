from abc import abstractmethod
from typing import List, Optional
from datetime import datetime
from src.core.domain.base_repository import BaseRepository
from src.modules.orders.domain.entities.order import Order, OrderStatus


class OrderRepository(BaseRepository[Order]):
    """Order repository interface (port)."""
    
    @abstractmethod
    async def get_by_buyer(self, buyer_id: int, skip: int = 0, limit: int = 100) -> List[Order]:
        """Get all orders for a buyer."""
        pass
    
    @abstractmethod
    async def get_by_store(self, store_id: int, skip: int = 0, limit: int = 100) -> List[Order]:
        """Get all orders for a store."""
        pass
    
    @abstractmethod
    async def get_by_qr_token(self, qr_token: str) -> Optional[Order]:
        """Get order by QR token."""
        pass
    
    @abstractmethod
    async def get_expired_reservations(self) -> List[Order]:
        """Get all expired reservations."""
        pass
    
    @abstractmethod
    async def update_status(self, order_id: int, status: OrderStatus) -> bool:
        """Update order status."""
        pass
