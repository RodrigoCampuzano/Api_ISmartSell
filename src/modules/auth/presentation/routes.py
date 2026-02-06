from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.infrastructure.database import get_db
from src.core.infrastructure.security import create_access_token, get_current_user
from src.modules.auth.application.dtos import (
    RegisterRequest,
    LoginRequest,
    UserResponse,
    TokenResponse
)
from src.modules.auth.application.use_cases.register import RegisterUseCase
from src.modules.auth.application.use_cases.login import LoginUseCase
from src.modules.auth.infrastructure.repositories.user_repository_impl import UserRepositoryImpl
from src.core.domain.exceptions import ValidationException, UnauthorizedException

router = APIRouter(prefix="/auth", tags=["Authentication"])


@router.post("/register", response_model=TokenResponse, status_code=status.HTTP_201_CREATED)
async def register(
    request: RegisterRequest,
    db: AsyncSession = Depends(get_db)
):
    """
    Register a new user.
    
    - **email**: User email address (unique)
    - **password**: Password (min 6 characters)
    - **role**: User role (BUYER, SELLER, ADMIN)
    - **username**: Optional username (unique)
    - **full_name**: Optional full name
    - **phone**: Optional phone number
    """
    try:
        user_repo = UserRepositoryImpl(db)
        use_case = RegisterUseCase(user_repo)
        
        user = await use_case.execute(
            email=request.email,
            password=request.password,
            role=request.role,
            username=request.username,
            full_name=request.full_name,
            phone=request.phone
        )
        
        # Create access token
        access_token = create_access_token(data={"sub": user.id})
        
        return TokenResponse(
            access_token=access_token,
            user=UserResponse.model_validate(user)
        )
    
    except ValidationException as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/login", response_model=TokenResponse)
async def login(
    request: LoginRequest,
    db: AsyncSession = Depends(get_db)
):
    """
    Login with email and password.
    
    Returns JWT access token for authentication.
    """
    try:
        user_repo = UserRepositoryImpl(db)
        use_case = LoginUseCase(user_repo)
        
        user = await use_case.execute(
            email=request.email,
            password=request.password
        )
        
        # Create access token
        access_token = create_access_token(data={"sub": user.id})
        
        return TokenResponse(
            access_token=access_token,
            user=UserResponse.model_validate(user)
        )
    
    except UnauthorizedException as e:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail=str(e))


@router.get("/me", response_model=UserResponse)
async def get_profile(
    current_user = Depends(get_current_user)
):
    """
    Get current authenticated user profile.
    
    Requires valid JWT token in Authorization header.
    """
    return UserResponse.model_validate(current_user)
