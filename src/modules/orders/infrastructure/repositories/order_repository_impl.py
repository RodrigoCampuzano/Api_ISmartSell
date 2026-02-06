from typing import Optional, List
from datetime import datetime
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from src.modules.orders.domain.entities.order import Order, OrderItem, OrderStatus, PaymentMethod
from src.modules.orders.domain.repositories.order_repository import OrderRepository
from src.modules.orders.infrastructure.models.order_model import (
    OrderModel,
    OrderItemModel,
    OrderStatusEnum,
    PaymentMethodEnum
)


class OrderRepositoryImpl(OrderRepository):
    """Implementation of OrderRepository using SQLAlchemy."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    def _to_entity(self, model: OrderModel) -> Order:
        """Convert ORM model to domain entity."""
        items = [
            OrderItem(
                id=item.id,
                product_id=item.product_id,
                quantity=item.quantity,
                unit_price=float(item.unit_price),
                total_price=float(item.total_price)
            )
            for item in model.items
        ]
        
        return Order(
            id=model.id,
            buyer_id=model.buyer_id,
            store_id=model.store_id,
            status=OrderStatus(model.status.value),
            total=float(model.total),
            subtotal=float(model.subtotal) if model.subtotal else None,
            shipping=float(model.shipping),
            payment_method=PaymentMethod(model.payment_method.value),
            qr_token=model.qr_token,
            reserved_until=model.reserved_until,
            items=items
        )
    
    async def create(self, entity: Order) -> Order:
        """Create a new order with items."""
        model = OrderModel(
            buyer_id=entity.buyer_id,
            store_id=entity.store_id,
            status=OrderStatusEnum(entity.status.value),
            total=entity.total,
            subtotal=entity.subtotal,
            shipping=entity.shipping,
            payment_method=PaymentMethodEnum(entity.payment_method.value),
            qr_token=entity.qr_token,
            reserved_until=entity.reserved_until
        )
        
        # Add items
        for item in entity.items:
            item_model = OrderItemModel(
                product_id=item.product_id,
                quantity=item.quantity,
                unit_price=item.unit_price,
                total_price=item.total_price
            )
            model.items.append(item_model)
        
        self.db.add(model)
        await self.db.flush()
        await self.db.refresh(model, ["items"])
        return self._to_entity(model)
    
    async def get_by_id(self, id: int) -> Optional[Order]:
        """Get order by ID with items."""
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .where(OrderModel.id == id)
        )
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_buyer(self, buyer_id: int, skip: int = 0, limit: int = 100) -> List[Order]:
        """Get all orders for a buyer."""
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .where(OrderModel.buyer_id == buyer_id)
            .offset(skip)
            .limit(limit)
            .order_by(OrderModel.created_at.desc())
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def get_by_store(self, store_id: int, skip: int = 0, limit: int = 100) -> List[Order]:
        """Get all orders for a store."""
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .where(OrderModel.store_id == store_id)
            .offset(skip)
            .limit(limit)
            .order_by(OrderModel.created_at.desc())
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def get_by_qr_token(self, qr_token: str) -> Optional[Order]:
        """Get order by QR token."""
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .where(OrderModel.qr_token == qr_token)
        )
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_expired_reservations(self) -> List[Order]:
        """Get all expired reservations."""
        now = datetime.utcnow()
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .where(
                OrderModel.status == OrderStatusEnum.RESERVED,
                OrderModel.reserved_until < now
            )
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def update_status(self, order_id: int, status: OrderStatus) -> bool:
        """Update order status."""
        result = await self.db.execute(select(OrderModel).where(OrderModel.id == order_id))
        model = result.scalar_one_or_none()
        
        if model:
            model.status = OrderStatusEnum(status.value)
            await self.db.flush()
            return True
        
        return False
    
    async def update(self, entity: Order) -> Order:
        """Update existing order."""
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .where(OrderModel.id == entity.id)
        )
        model = result.scalar_one_or_none()
        
        if model:
            model.status = OrderStatusEnum(entity.status.value)
            model.qr_token = entity.qr_token
            model.payment_method = PaymentMethodEnum(entity.payment_method.value)
            await self.db.flush()
            await self.db.refresh(model, ["items"])
            return self._to_entity(model)
        
        return None
    
    async def delete(self, id: int) -> bool:
        """Delete order by ID."""
        result = await self.db.execute(select(OrderModel).where(OrderModel.id == id))
        model = result.scalar_one_or_none()
        
        if model:
            await self.db.delete(model)
            await self.db.flush()
            return True
        
        return False
    
    async def list(self, skip: int = 0, limit: int = 100) -> List[Order]:
        """List orders with pagination."""
        result = await self.db.execute(
            select(OrderModel)
            .options(selectinload(OrderModel.items))
            .offset(skip)
            .limit(limit)
            .order_by(OrderModel.created_at.desc())
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
