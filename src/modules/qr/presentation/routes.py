from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.infrastructure.database import get_db
from src.modules.qr.application.dtos import QRValidateRequest, QRValidateResponse
from src.modules.qr.application.use_cases.validate_qr import ValidateQRUseCase
from src.modules.orders.infrastructure.repositories.order_repository_impl import OrderRepositoryImpl

router = APIRouter(prefix="/qr", tags=["QR Validation"])


@router.post("/validate", response_model=QRValidateResponse)
async def validate_qr(
    request: QRValidateRequest,
    db: AsyncSession = Depends(get_db)
):
    """
    Validate QR token and deliver order.
    
    This endpoint is typically called by the seller when scanning
    the buyer's QR code at pickup.
    
    Returns validation result and updates order status to DELIVERED if valid.
    """
    order_repo = OrderRepositoryImpl(db)
    use_case = ValidateQRUseCase(order_repo)
    
    result = await use_case.execute(request.qr_token)
    
    return QRValidateResponse(**result)
