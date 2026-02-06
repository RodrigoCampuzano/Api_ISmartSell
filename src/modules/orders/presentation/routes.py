from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List

from src.core.infrastructure.database import get_db
from src.core.infrastructure.security import get_current_user
from src.modules.orders.application.dtos import (
    CreateOrderRequest,
    OrderResponse,
    OrderItemResponse
)
from src.modules.orders.application.use_cases.create_order import CreateOrderUseCase
from src.modules.orders.application.use_cases.cancel_order import CancelOrderUseCase
from src.modules.orders.domain.entities.order import OrderStatus
from src.modules.orders.infrastructure.repositories.order_repository_impl import OrderRepositoryImpl
from src.modules.products.infrastructure.repositories.product_repository_impl import ProductRepositoryImpl
from src.modules.stores.infrastructure.repositories.store_repository_impl import StoreRepositoryImpl
from src.core.domain.exceptions import ValidationException, BusinessRuleViolationException
from src.modules.auth.domain.entities.user import UserRole

router = APIRouter(prefix="/orders", tags=["Orders"])


@router.post("", response_model=OrderResponse, status_code=status.HTTP_201_CREATED)
async def create_order(
    request: CreateOrderRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Create a new order.
    
    - **store_id**: Target store ID
    - **items**: List of products with quantities
    - **payment_method**: ONLINE, CASH, or NONE (for reservation)
    """
    try:
        order_repo = OrderRepositoryImpl(db)
        product_repo = ProductRepositoryImpl(db)
        
        use_case = CreateOrderUseCase(order_repo, product_repo)
        
        order = await use_case.execute(
            buyer_id=current_user.id,
            store_id=request.store_id,
            items=[{"product_id": item.product_id, "quantity": item.quantity} for item in request.items],
            payment_method=request.payment_method
        )
        
        # Convert items
        items_response = [
            OrderItemResponse(
                id=item.id,
                product_id=item.product_id,
                quantity=item.quantity,
                unit_price=item.unit_price,
                total_price=item.total_price
            )
            for item in order.items
        ]
        
        response = OrderResponse.model_validate(order)
        response.items = items_response
        return response
    
    except (ValidationException, BusinessRuleViolationException) as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/{order_id}", response_model=OrderResponse)
async def get_order(
    order_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get order by ID (buyer or seller access)."""
    order_repo = OrderRepositoryImpl(db)
    order = await order_repo.get_by_id(order_id)
    
    if not order:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Order {order_id} not found"
        )
    
    # Check access: buyer owns order OR seller owns store
    if order.buyer_id != current_user.id:
        if current_user.role == UserRole.SELLER:
            store_repo = StoreRepositoryImpl(db)
            store = await store_repo.get_by_id(order.store_id)
            if not store or store.seller_id != current_user.id:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail="Access denied"
                )
        else:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied"
            )
    
    items_response = [
        OrderItemResponse.model_validate(item)
        for item in order.items
    ]
    
    response = OrderResponse.model_validate(order)
    response.items = items_response
    return response


@router.get("/users/{user_id}/orders", response_model=List[OrderResponse])
async def get_user_orders(
    user_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get order history for a user (own orders only)."""
    if user_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only view your own orders"
        )
    
    order_repo = OrderRepositoryImpl(db)
    orders = await order_repo.get_by_buyer(user_id)
    
    return [
        OrderResponse.model_validate(order) if not order.items 
        else OrderResponse(
            **{k: v for k, v in order.__dict__.items() if k != 'items'},
            items=[OrderItemResponse.model_validate(item) for item in order.items]
        )
        for order in orders
    ]


@router.put("/{order_id}/cancel", status_code=status.HTTP_200_OK)
async def cancel_order(
    order_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Cancel an order (restores stock).
    
    Only the buyer can cancel their own orders.
    """
    try:
        order_repo = OrderRepositoryImpl(db)
        product_repo = ProductRepositoryImpl(db)
        
        use_case = CancelOrderUseCase(order_repo, product_repo)
        success = await use_case.execute(order_id, current_user.id)
        
        if success:
            return {"message": "Order cancelled successfully"}
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Failed to cancel order"
            )
    
    except (ValidationException, BusinessRuleViolationException) as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.put("/{order_id}/mark-ready", status_code=status.HTTP_200_OK)
async def mark_order_ready(
    order_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Mark order as ready for pickup (Seller only).
    
    Only the store owner can mark orders as ready.
    """
    order_repo = OrderRepositoryImpl(db)
    order = await order_repo.get_by_id(order_id)
    
    if not order:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Order {order_id} not found"
        )
    
    # Check seller owns the store
    store_repo = StoreRepositoryImpl(db)
    store = await store_repo.get_by_id(order.store_id)
    
    if not store or store.seller_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only mark orders from your own stores as ready"
        )
    
    if not order.can_mark_ready():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Order cannot be marked as ready (status: {order.status})"
        )
    
    success = await order_repo.update_status(order_id, OrderStatus.READY)
    
    if success:
        return {"message": "Order marked as ready"}
    else:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Failed to update order status"
        )
