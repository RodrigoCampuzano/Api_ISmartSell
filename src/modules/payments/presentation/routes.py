from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.infrastructure.database import get_db
from src.core.infrastructure.security import get_current_user
from src.modules.payments.application.dtos import PaymentRequest, PaymentResponse
from src.modules.payments.domain.entities.payment import Payment, PaymentStatus
from src.modules.payments.infrastructure.repositories.payment_repository_impl import PaymentRepositoryImpl
from src.modules.payments.application.use_cases.process_payment import ProcessPaymentUseCase
from src.modules.orders.infrastructure.repositories.order_repository_impl import OrderRepositoryImpl
from src.core.domain.exceptions import ValidationException

router = APIRouter(prefix="/payments", tags=["Payments"])


@router.post("/{order_id}/pay", response_model=PaymentResponse, status_code=status.HTTP_201_CREATED)
async def initiate_payment(
    order_id: int,
    request: PaymentRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Initiate payment for an order.
    
    Creates payment record. In production, this would redirect to payment gateway.
    """
    order_repo = OrderRepositoryImpl(db)
    order = await order_repo.get_by_id(order_id)
    
    if not order:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Order {order_id} not found"
        )
    
    if order.buyer_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only pay for your own orders"
        )
    
    # Check if payment already exists
    payment_repo = PaymentRepositoryImpl(db)
    existing = await payment_repo.get_by_order(order_id)
    
    if existing:
        return PaymentResponse.model_validate(existing)
    
    # Create payment
    payment = Payment(
        order_id=order_id,
        amount=order.total,
        provider=request.provider,
        status=PaymentStatus.CREATED
    )
    
    created = await payment_repo.create(payment)
    
    # In production, here you would:
    # 1. Call payment provider API (Stripe, PayU, etc.)
    # 2. Return payment URL or token for frontend
    
    return PaymentResponse.model_validate(created)


@router.post("/webhook", status_code=status.HTTP_200_OK)
async def payment_webhook(
    db: AsyncSession = Depends(get_db)
):
    """
    Webhook endpoint for payment provider notifications.
    
    WARNING: In production, MUST validate provider signature!
    For demo, this endpoint simulates successful payment.
    """
    # In production:
    # 1. Validate webhook signature from provider
    # 2. Extract payment_id from webhook payload
    # 3. Call ProcessPaymentUseCase
    
    # For demo purposes, this is a simplified version
    # You would extract payment_id from the webhook payload
    
    return {"message": "Webhook received. Use POST /payments/{payment_id}/complete for demo."}


@router.post("/{payment_id}/complete", response_model=PaymentResponse)
async def complete_payment(
    payment_id: int,
    db: AsyncSession = Depends(get_db)
):
    """
    Complete a payment (Demo endpoint - simulates webhook).
    
    In production, this would be called by the webhook handler.
    Generates QR token and records 1% commission.
    """
    try:
        payment_repo = PaymentRepositoryImpl(db)
        order_repo = OrderRepositoryImpl(db)
        
        use_case = ProcessPaymentUseCase(payment_repo, order_repo)
        success = await use_case.execute(payment_id)
        
        if success:
            payment = await payment_repo.get_by_id(payment_id)
            return PaymentResponse.model_validate(payment)
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Failed to process payment"
            )
    
    except ValidationException as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/{payment_id}", response_model=PaymentResponse)
async def get_payment(
    payment_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get payment details."""
    payment_repo = PaymentRepositoryImpl(db)
    payment = await payment_repo.get_by_id(payment_id)
    
    if not payment:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Payment {payment_id} not found"
        )
    
    # Check access
    order_repo = OrderRepositoryImpl(db)
    order = await order_repo.get_by_id(payment.order_id)
    
    if order.buyer_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Access denied"
        )
    
    return PaymentResponse.model_validate(payment)
