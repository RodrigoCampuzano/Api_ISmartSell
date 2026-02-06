from typing import Optional
from src.core.domain.base_entity import BaseEntity


class Store(BaseEntity):
    """Store domain entity."""
    
    def __init__(
        self,
        seller_id: int,
        name: str,
        slug: str,
        description: Optional[str] = None,
        address: Optional[str] = None,
        lat: Optional[float] = None,
        lng: Optional[float] = None,
        id: Optional[int] = None
    ):
        self.id = id
        self.seller_id = seller_id
        self.name = name
        self.slug = slug
        self.description = description
        self.address = address
        self.lat = lat
        self.lng = lng
