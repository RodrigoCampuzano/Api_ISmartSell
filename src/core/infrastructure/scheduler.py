"""Background tasks for the ISmartSell API."""
import asyncio
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from sqlalchemy.ext.asyncio import AsyncSession
from src.core.infrastructure.database import AsyncSessionLocal
from src.modules.orders.infrastructure.repositories.order_repository_impl import OrderRepositoryImpl
from src.modules.products.infrastructure.repositories.product_repository_impl import ProductRepositoryImpl
from src.modules.orders.domain.entities.order import OrderStatus
import logging

logger = logging.getLogger(__name__)


async def cancel_expired_reservations():
    """
    Background task to cancel expired order reservations.
    
    Runs periodically to find orders with status RESERVED where
    reserved_until < now, then cancels them and restores stock.
    """
    async with AsyncSessionLocal() as db:
        try:
            order_repo = OrderRepositoryImpl(db)
            product_repo = ProductRepositoryImpl(db)
            
            # Get expired reservations
            expired_orders = await order_repo.get_expired_reservations()
            
            logger.info(f"Found {len(expired_orders)} expired reservations")
            
            for order in expired_orders:
                # Restore stock for each item
                for item in order.items:
                    product = await product_repo.get_by_id(item.product_id)
                    if product:
                        product.restore_stock(item.quantity)
                        await product_repo.update(product)
                        logger.info(f"Restored {item.quantity} units of product {product.id}")
                
                # Cancel order
                await order_repo.update_status(order.id, OrderStatus.CANCELLED)
                logger.info(f"Cancelled expired order {order.id}")
            
            await db.commit()
        
        except Exception as e:
            logger.error(f"Error cancelling expired reservations: {e}")
            await db.rollback()


def start_scheduler():
    """Initialize and start the background task scheduler."""
    scheduler = AsyncIOScheduler()
    
    # Run every 5 minutes
    scheduler.add_job(
        cancel_expired_reservations,
        'interval',
        minutes=5,
        id='cancel_expired_reservations'
    )
    
    scheduler.start()
    logger.info("Background scheduler started")
    
    return scheduler
