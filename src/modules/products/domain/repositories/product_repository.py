from abc import abstractmethod
from typing import List
from src.core.domain.base_repository import BaseRepository
from src.modules.products.domain.entities.product import Product


class ProductRepository(BaseRepository[Product]):
    """Product repository interface (port)."""
    
    @abstractmethod
    async def get_by_store(self, store_id: int, skip: int = 0, limit: int = 100) -> List[Product]:
        """Get all products for a store."""
        pass
    
    @abstractmethod
    async def search(self, query: str, skip: int = 0, limit: int = 100) -> List[Product]:
        """Search products by name or description."""
        pass
