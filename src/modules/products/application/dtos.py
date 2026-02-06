from pydantic import BaseModel, Field
from typing import Optional


class ProductRequest(BaseModel):
    """Request model for creating/updating product."""
    name: str = Field(..., min_length=1)
    sku: Optional[str] = None
    description: Optional[str] = None
    price: float = Field(..., gt=0)
    stock: int = Field(0, ge=0)
    image_url: Optional[str] = None
    active: bool = True


class ProductResponse(BaseModel):
    """Response model for product data."""
    id: int
    store_id: int
    name: str
    sku: Optional[str]
    description: Optional[str]
    price: float
    stock: int
    image_url: Optional[str]
    active: bool
    
    class Config:
        from_attributes = True
