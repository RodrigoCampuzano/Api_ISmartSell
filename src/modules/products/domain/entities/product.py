from typing import Optional
from src.core.domain.base_entity import BaseEntity


class Product(BaseEntity):
    """Product domain entity."""
    
    def __init__(
        self,
        store_id: int,
        name: str,
        price: float,
        stock: int = 0,
        sku: Optional[str] = None,
        description: Optional[str] = None,
        image_url: Optional[str] = None,
        active: bool = True,
        id: Optional[int] = None
    ):
        self.id = id
        self.store_id = store_id
        self.name = name
        self.sku = sku
        self.description = description
        self.price = price
        self.stock = stock
        self.image_url = image_url
        self.active = active
    
    def is_available(self) -> bool:
        """Check if product is available for purchase."""
        return self.active and self.stock > 0
    
    def reserve_stock(self, quantity: int) -> bool:
        """Reserve stock for order."""
        if self.stock >= quantity:
            self.stock -= quantity
            return True
        return False
    
    def restore_stock(self, quantity: int):
        """Restore stock when order is cancelled."""
        self.stock += quantity
