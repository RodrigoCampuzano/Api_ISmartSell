from datetime import datetime, timedelta
from typing import Optional
from jose import JWTError, jwt
from passlib.context import CryptContext
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sqlalchemy.ext.asyncio import AsyncSession

from src.config import settings
from src.core.infrastructure.database import get_db

# Password hashing
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

# JWT Bearer token
security = HTTPBearer()


def hash_password(password: str) -> str:
    """
    Hash a password using bcrypt.
    Truncate to 72 bytes if necessary (bcrypt limitation).
    """
    # Truncate password to 72 bytes (bcrypt limitation)
    password_bytes = password.encode('utf-8')
    if len(password_bytes) > 72:
        password = password_bytes[:72].decode('utf-8', errors='ignore')
    return pwd_context.hash(password)


def verify_password(plain_password: str, hashed_password: str) -> bool:
    """Verify a password against its hash."""
    return pwd_context.verify(plain_password, hashed_password)


def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    """Create a JWT access token."""
    to_encode = data.copy()
    
    # Convert 'sub' to string if it's an int (JWT requires string)
    if 'sub' in to_encode and isinstance(to_encode['sub'], int):
        to_encode['sub'] = str(to_encode['sub'])
    
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES)
    
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)
    return encoded_jwt


def decode_access_token(token: str) -> dict:
    """Decode and verify a JWT access token."""
    print(f"DEBUG decode_access_token: Token length: {len(token)}")
    print(f"DEBUG decode_access_token: JWT_SECRET_KEY: {settings.JWT_SECRET_KEY[:20]}...")
    print(f"DEBUG decode_access_token: JWT_ALGORITHM: {settings.JWT_ALGORITHM}")
    
    try:
        payload = jwt.decode(token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM])
        print(f"DEBUG decode_access_token: Successfully decoded payload: {payload}")
        return payload
    except JWTError as e:
        print(f"DEBUG decode_access_token: JWTError details: {type(e).__name__}: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )
    except Exception as e:
        print(f"DEBUG decode_access_token: Unexpected error: {type(e).__name__}: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )


async def get_current_user_id(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> int:
    """Extract user ID from JWT token."""
    token = credentials.credentials
    print(f"DEBUG: Received token: {token[:50]}...")  # Print first 50 chars
    
    try:
        payload = decode_access_token(token)
        print(f"DEBUG: Decoded payload: {payload}")
    except Exception as e:
        print(f"DEBUG: Error decoding token: {e}")
        raise
    
    user_id = payload.get("sub")
    print(f"DEBUG: Extracted user_id: {user_id} (type: {type(user_id)})")
    
    if user_id is None:
        print("DEBUG: user_id is None!")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    # Convert to int if it's a string
    try:
        user_id_int = int(user_id)
        print(f"DEBUG: Converted user_id to int: {user_id_int}")
        return user_id_int
    except (ValueError, TypeError) as e:
        print(f"DEBUG: Error converting user_id to int: {e}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token payload",
            headers={"WWW-Authenticate": "Bearer"},
        )


async def get_current_user(
    user_id: int = Depends(get_current_user_id),
    db: AsyncSession = Depends(get_db)
):
    """Get current authenticated user from database."""
    from src.modules.auth.infrastructure.repositories.user_repository_impl import UserRepositoryImpl
    
    user_repo = UserRepositoryImpl(db)
    user = await user_repo.get_by_id(user_id)
    
    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="User not found"
        )
    
    return user
