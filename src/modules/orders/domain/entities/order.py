from enum import Enum
from typing import Optional, List
from datetime import datetime
from src.core.domain.base_entity import BaseEntity


class OrderStatus(str, Enum):
    """Order status states."""
    PENDING = "PENDING"
    RESERVED = "RESERVED"
    PAID = "PAID"
    READY = "READY"
    DELIVERED = "DELIVERED"
    CANCELLED = "CANCELLED"


class PaymentMethod(str, Enum):
    """Payment methods."""
    ONLINE = "ONLINE"
    CASH = "CASH"
    NONE = "NONE"


class OrderItem:
    """Order item value object."""
    
    def __init__(
        self,
        product_id: int,
        quantity: int,
        unit_price: float,
        total_price: Optional[float] = None,
        id: Optional[int] = None
    ):
        self.id = id
        self.product_id = product_id
        self.quantity = quantity
        self.unit_price = unit_price
        self.total_price = total_price or (unit_price * quantity)


class Order(BaseEntity):
    """Order domain entity."""
    
    def __init__(
        self,
        buyer_id: int,
        store_id: int,
        total: float,
        status: OrderStatus = OrderStatus.PENDING,
        subtotal: Optional[float] = None,
        shipping: float = 0.0,
        payment_method: PaymentMethod = PaymentMethod.NONE,
        qr_token: Optional[str] = None,
        reserved_until: Optional[datetime] = None,
        items: Optional[List[OrderItem]] = None,
        id: Optional[int] = None
    ):
        self.id = id
        self.buyer_id = buyer_id
        self.store_id = store_id
        self.status = status
        self.total = total
        self.subtotal = subtotal
        self.shipping = shipping
        self.payment_method = payment_method
        self.qr_token = qr_token
        self.reserved_until = reserved_until
        self.items = items or []
    
    def can_cancel(self) -> bool:
        """Check if order can be cancelled."""
        return self.status in [OrderStatus.PENDING, OrderStatus.RESERVED, OrderStatus.PAID]
    
    def can_mark_ready(self) -> bool:
        """Check if order can be marked as ready."""
        return self.status == OrderStatus.PAID
    
    def can_deliver(self) -> bool:
        """Check if order can be delivered."""
        return self.status in [OrderStatus.PAID, OrderStatus.READY]
    
    def is_expired(self) -> bool:
        """Check if reservation has expired."""
        if self.status == OrderStatus.RESERVED and self.reserved_until:
            return datetime.utcnow() > self.reserved_until
        return False
