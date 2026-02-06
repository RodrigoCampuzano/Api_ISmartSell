from abc import abstractmethod
from typing import Optional, List
from src.core.domain.base_repository import BaseRepository
from src.modules.stores.domain.entities.store import Store
from src.modules.stores.domain.entities.delivery_point import DeliveryPoint


class StoreRepository(BaseRepository[Store]):
    """Store repository interface (port)."""
    
    @abstractmethod
    async def get_by_slug(self, slug: str) -> Optional[Store]:
        """Get store by slug."""
        pass
    
    @abstractmethod
    async def get_by_seller(self, seller_id: int) -> List[Store]:
        """Get all stores for a seller."""
        pass
    
    @abstractmethod
    async def search(self, query: str, skip: int = 0, limit: int = 100) -> List[Store]:
        """Search stores by name or description."""
        pass
    
    @abstractmethod
    async def get_nearby(self, lat: float, lng: float, radius_km: float = 10.0) -> List[Store]:
        """Get stores within radius of coordinates."""
        pass
    
    @abstractmethod
    async def add_delivery_point(self, point: DeliveryPoint) -> DeliveryPoint:
        """Add delivery point to store."""
        pass
    
    @abstractmethod
    async def get_delivery_points(self, store_id: int) -> List[DeliveryPoint]:
        """Get all delivery points for a store."""
        pass
