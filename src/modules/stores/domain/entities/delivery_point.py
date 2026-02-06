from typing import Optional
from src.core.domain.base_entity import BaseEntity


class DeliveryPoint(BaseEntity):
    """Delivery point domain entity."""
    
    def __init__(
        self,
        store_id: int,
        name: str,
        address: Optional[str] = None,
        lat: Optional[float] = None,
        lng: Optional[float] = None,
        id: Optional[int] = None
    ):
        self.id = id
        self.store_id = store_id
        self.name = name
        self.address = address
        self.lat = lat
        self.lng = lng
