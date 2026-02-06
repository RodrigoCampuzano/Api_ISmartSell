import uuid
from src.modules.payments.domain.entities.payment import Payment, PaymentStatus
from src.modules.payments.domain.repositories.payment_repository import PaymentRepository
from src.modules.orders.domain.repositories.order_repository import OrderRepository
from src.modules.orders.domain.entities.order import OrderStatus
from src.core.domain.exceptions import ValidationException
from src.config import settings


class ProcessPaymentUseCase:
    """Use case for processing payment completion."""
    
    def __init__(
        self,
        payment_repository: PaymentRepository,
        order_repository: OrderRepository
    ):
        self.payment_repository = payment_repository
        self.order_repository = order_repository
    
    async def execute(self, payment_id: int) -> bool:
        """
        Process payment completion.
        Updates order status to PAID, generates QR token, and records commission.
        """
        
        payment = await self.payment_repository.get_by_id(payment_id)
        
        if not payment:
            raise ValidationException(f"Payment {payment_id} not found")
        
        if payment.status != PaymentStatus.CREATED:
            raise ValidationException(f"Payment already processed (status: {payment.status})")
        
        # Get associated order
        order = await self.order_repository.get_by_id(payment.order_id)
        
        if not order:
            raise ValidationException(f"Order {payment.order_id} not found")
        
        # Calculate commission
        payment.calculate_commission(settings.PLATFORM_COMMISSION_RATE)
        payment.status = PaymentStatus.COMPLETED
        
        # Update payment
        await self.payment_repository.update(payment)
        
        # Record platform revenue
        await self.payment_repository.record_platform_revenue(
            payment.id,
            payment.platform_commission
        )
        
        # Generate QR token and update order
        qr_token = str(uuid.uuid4())
        order.qr_token = qr_token
        order.status = OrderStatus.PAID
        await self.order_repository.update(order)
        
        return True
