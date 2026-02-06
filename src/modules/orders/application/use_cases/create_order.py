from datetime import datetime, timedelta
from typing import List
from src.modules.orders.domain.entities.order import Order, OrderItem, OrderStatus, PaymentMethod
from src.modules.orders.domain.repositories.order_repository import OrderRepository
from src.modules.products.domain.repositories.product_repository import ProductRepository
from src.core.domain.exceptions import ValidationException, BusinessRuleViolationException
from src.config import settings


class CreateOrderUseCase:
    """Use case for creating an order."""
    
    def __init__(
        self,
        order_repository: OrderRepository,
        product_repository: ProductRepository
    ):
        self.order_repository = order_repository
        self.product_repository = product_repository
    
    async def execute(
        self,
        buyer_id: int,
        store_id: int,
        items: List[dict],
        payment_method: PaymentMethod = PaymentMethod.NONE
    ) -> Order:
        """Create a new order with items."""
        
        if not items:
            raise ValidationException("Order must have at least one item")
        
        # Validate products and calculate totals
        order_items = []
        subtotal = 0.0
        
        for item_data in items:
            product = await self.product_repository.get_by_id(item_data["product_id"])
            
            if not product:
                raise ValidationException(f"Product {item_data['product_id']} not found")
            
            if product.store_id != store_id:
                raise ValidationException(f"Product {product.id} does not belong to store {store_id}")
            
            if not product.is_available():
                raise BusinessRuleViolationException(f"Product {product.name} is not available")
            
            if product.stock < item_data["quantity"]:
                raise BusinessRuleViolationException(
                    f"Insufficient stock for {product.name}. Available: {product.stock}, Requested: {item_data['quantity']}"
                )
            
            # Reserve stock
            product.reserve_stock(item_data["quantity"])
            await self.product_repository.update(product)
            
            # Create order item
            unit_price = product.price
            total_price = unit_price * item_data["quantity"]
            order_items.append(OrderItem(
                product_id=product.id,
                quantity=item_data["quantity"],
                unit_price=unit_price,
                total_price=total_price
            ))
            
            subtotal += total_price
        
        # Calculate total
        shipping = 0.0  # Can be customized
        total = subtotal + shipping
        
        # Set status and reservation
        status = OrderStatus.RESERVED if payment_method == PaymentMethod.NONE else OrderStatus.PENDING
        reserved_until = None
        
        if status == OrderStatus.RESERVED:
            reserved_until = datetime.utcnow() + timedelta(minutes=settings.RESERVATION_TIMEOUT_MINUTES)
        
        # Create order
        order = Order(
            buyer_id=buyer_id,
            store_id=store_id,
            status=status,
            total=total,
            subtotal=subtotal,
            shipping=shipping,
            payment_method=payment_method,
            reserved_until=reserved_until,
            items=order_items
        )
        
        return await self.order_repository.create(order)
