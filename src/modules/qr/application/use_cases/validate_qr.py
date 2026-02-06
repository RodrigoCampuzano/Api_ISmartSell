from src.modules.orders.domain.repositories.order_repository import OrderRepository
from src.modules.orders.domain.entities.order import OrderStatus


class ValidateQRUseCase:
    """Use case for validating QR token and delivering order."""
    
    def __init__(self, order_repository: OrderRepository):
        self.order_repository = order_repository
    
    async def execute(self, qr_token: str) -> dict:
        """
        Validate QR token and mark order as delivered.
        
        Returns validation result with order details.
        """
        
        # Find order by QR token
        order = await self.order_repository.get_by_qr_token(qr_token)
        
        if not order:
            return {
                "valid": False,
                "message": "Invalid QR code"
            }
        
        # Check if order can be delivered
        if not order.can_deliver():
            return {
                "valid": False,
                "order_id": order.id,
                "store_id": order.store_id,
                "status": order.status.value,
                "message": f"Order cannot be delivered (current status: {order.status.value})"
            }
        
        # Mark as delivered
        success = await self.order_repository.update_status(order.id, OrderStatus.DELIVERED)
        
        if success:
            return {
                "valid": True,
                "order_id": order.id,
                "store_id": order.store_id,
                "status": OrderStatus.DELIVERED.value,
                "message": "Order validated and marked as delivered"
            }
        else:
            return {
                "valid": False,
                "order_id": order.id,
                "store_id": order.store_id,
                "status": order.status.value,
                "message": "Failed to update order status"
            }
