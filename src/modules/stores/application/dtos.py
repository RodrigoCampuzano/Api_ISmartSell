from pydantic import BaseModel, Field
from typing import Optional


class StoreRequest(BaseModel):
    """Request model for creating/updating a store."""
    name: str = Field(..., min_length=1)
    slug: str = Field(..., min_length=1)
    description: Optional[str] = None
    address: Optional[str] = None
    lat: Optional[float] = Field(None, ge=-90, le=90)
    lng: Optional[float] = Field(None, ge=-180, le=180)


class StoreResponse(BaseModel):
    """Response model for store data."""
    id: int
    seller_id: int
    name: str
    slug: str
    description: Optional[str]
    address: Optional[str]
    lat: Optional[float]
    lng: Optional[float]
    
    class Config:
        from_attributes = True


class DeliveryPointRequest(BaseModel):
    """Request model for adding delivery point."""
    name: str
    address: Optional[str] = None
    lat: Optional[float] = Field(None, ge=-90, le=90)
    lng: Optional[float] = Field(None, ge=-180, le=180)


class DeliveryPointResponse(BaseModel):
    """Response model for delivery point."""
    id: int
    store_id: int
    name: str
    address: Optional[str]
    lat: Optional[float]
    lng: Optional[float]
    
    class Config:
        from_attributes = True
