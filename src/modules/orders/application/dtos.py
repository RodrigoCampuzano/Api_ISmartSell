from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime
from src.modules.orders.domain.entities.order import OrderStatus, PaymentMethod


class OrderItemRequest(BaseModel):
    """Request model for order item."""
    product_id: int
    quantity: int = Field(..., gt=0)


class CreateOrderRequest(BaseModel):
    """Request model for creating an order."""
    store_id: int
    items: List[OrderItemRequest] = Field(..., min_length=1)
    payment_method: PaymentMethod = PaymentMethod.NONE
    shipping_point_id: Optional[int] = None


class OrderItemResponse(BaseModel):
    """Response model for order item."""
    id: int
    product_id: int
    quantity: int
    unit_price: float
    total_price: float
    
    class Config:
        from_attributes = True


class OrderResponse(BaseModel):
    """Response model for order."""
    id: int
    buyer_id: int
    store_id: int
    status: OrderStatus
    total: float
    subtotal: Optional[float]
    shipping: float
    payment_method: PaymentMethod
    qr_token: Optional[str]
    reserved_until: Optional[datetime]
    items: List[OrderItemResponse] = []
    created_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True
