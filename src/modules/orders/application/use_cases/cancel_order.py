from src.modules.orders.domain.entities.order import OrderStatus
from src.modules.orders.domain.repositories.order_repository import OrderRepository
from src.modules.products.domain.repositories.product_repository import ProductRepository
from src.core.domain.exceptions import ValidationException, BusinessRuleViolationException


class CancelOrderUseCase:
    """Use case for cancelling an order."""
    
    def __init__(
        self,
        order_repository: OrderRepository,
        product_repository: ProductRepository
    ):
        self.order_repository = order_repository
        self.product_repository = product_repository
    
    async def execute(self, order_id: int, user_id: int) -> bool:
        """Cancel an order and restore stock."""
        
        order = await self.order_repository.get_by_id(order_id)
        
        if not order:
            raise ValidationException(f"Order {order_id} not found")
        
        if not order.can_cancel():
            raise BusinessRuleViolationException(f"Order cannot be cancelled (status: {order.status})")
        
        # Only buyer can cancel their own orders
        if order.buyer_id != user_id:
            raise BusinessRuleViolationException("You can only cancel your own orders")
        
        # Restore stock for all items
        for item in order.items:
            product = await self.product_repository.get_by_id(item.product_id)
            if product:
                product.restore_stock(item.quantity)
                await self.product_repository.update(product)
        
        # Update order status
        return await self.order_repository.update_status(order_id, OrderStatus.CANCELLED)
