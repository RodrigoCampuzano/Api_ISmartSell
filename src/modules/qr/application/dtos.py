from pydantic import BaseModel


class QRValidateRequest(BaseModel):
    """Request model for QR validation."""
    qr_token: str


class QRValidateResponse(BaseModel):
    """Response model for QR validation."""
    valid: bool
    order_id: int | None = None
    store_id: int | None = None
    status: str | None = None
    message: str
