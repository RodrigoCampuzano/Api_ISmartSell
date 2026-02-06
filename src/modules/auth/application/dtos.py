from pydantic import BaseModel, EmailStr, Field
from typing import Optional
from src.modules.auth.domain.entities.user import UserRole


class RegisterRequest(BaseModel):
    """Request model for user registration."""
    email: EmailStr
    password: str = Field(..., min_length=6)
    role: UserRole = UserRole.BUYER
    username: Optional[str] = None
    full_name: Optional[str] = None
    phone: Optional[str] = None


class LoginRequest(BaseModel):
    """Request model for user login."""
    email: EmailStr
    password: str


class UserResponse(BaseModel):
    """Response model for user data."""
    id: int
    email: str
    username: Optional[str]
    role: UserRole
    full_name: Optional[str]
    phone: Optional[str]
    
    class Config:
        from_attributes = True


class TokenResponse(BaseModel):
    """Response model for authentication token."""
    access_token: str
    token_type: str = "bearer"
    user: UserResponse
