from typing import Optional, List
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from src.modules.payments.domain.entities.payment import Payment, PaymentStatus
from src.modules.payments.domain.repositories.payment_repository import PaymentRepository
from src.modules.payments.infrastructure.models.payment_model import (
    PaymentModel,
    PlatformRevenueModel,
    PaymentStatusEnum
)


class PaymentRepositoryImpl(PaymentRepository):
    """Implementation of PaymentRepository using SQLAlchemy."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    def _to_entity(self, model: PaymentModel) -> Payment:
        """Convert ORM model to domain entity."""
        return Payment(
            id=model.id,
            order_id=model.order_id,
            amount=float(model.amount),
            provider=model.provider,
            provider_fee=float(model.provider_fee),
            platform_commission=float(model.platform_commission),
            status=PaymentStatus(model.status.value)
        )
    
    async def create(self, entity: Payment) -> Payment:
        """Create a new payment."""
        model = PaymentModel(
            order_id=entity.order_id,
            amount=entity.amount,
            provider=entity.provider,
            provider_fee=entity.provider_fee,
            platform_commission=entity.platform_commission,
            status=PaymentStatusEnum(entity.status.value)
        )
        self.db.add(model)
        await self.db.flush()
        await self.db.refresh(model)
        return self._to_entity(model)
    
    async def get_by_id(self, id: int) -> Optional[Payment]:
        """Get payment by ID."""
        result = await self.db.execute(select(PaymentModel).where(PaymentModel.id == id))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_order(self, order_id: int) -> Optional[Payment]:
        """Get payment by order ID."""
        result = await self.db.execute(select(PaymentModel).where(PaymentModel.order_id == order_id))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def update(self, entity: Payment) -> Payment:
        """Update existing payment."""
        result = await self.db.execute(select(PaymentModel).where(PaymentModel.id == entity.id))
        model = result.scalar_one_or_none()
        
        if model:
            model.status = PaymentStatusEnum(entity.status.value)
            model.platform_commission = entity.platform_commission
            model.provider_fee = entity.provider_fee
            await self.db.flush()
            await self.db.refresh(model)
            return self._to_entity(model)
        
        return None
    
    async def delete(self, id: int) -> bool:
        """Delete payment by ID."""
        result = await self.db.execute(select(PaymentModel).where(PaymentModel.id == id))
        model = result.scalar_one_or_none()
        
        if model:
            await self.db.delete(model)
            await self.db.flush()
            return True
        
        return False
    
    async def list(self, skip: int = 0, limit: int = 100) -> List[Payment]:
        """List payments with pagination."""
        result = await self.db.execute(select(PaymentModel).offset(skip).limit(limit))
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def record_platform_revenue(self, payment_id: int, amount: float) -> bool:
        """Record platform revenue/commission."""
        revenue = PlatformRevenueModel(
            payment_id=payment_id,
            amount=amount
        )
        self.db.add(revenue)
        await self.db.flush()
        return True
