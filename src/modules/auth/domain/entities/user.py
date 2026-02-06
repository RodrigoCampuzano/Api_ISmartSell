from enum import Enum
from typing import Optional
from src.core.domain.base_entity import BaseEntity


class UserRole(str, Enum):
    """User roles in the system."""
    BUYER = "BUYER"
    SELLER = "SELLER"
    ADMIN = "ADMIN"


class User(BaseEntity):
    """User domain entity."""
    
    def __init__(
        self,
        email: str,
        password_hash: str,
        role: UserRole = UserRole.BUYER,
        username: Optional[str] = None,
        full_name: Optional[str] = None,
        phone: Optional[str] = None,
        id: Optional[int] = None
    ):
        self.id = id
        self.email = email
        self.username = username
        self.password_hash = password_hash
        self.role = role
        self.full_name = full_name
        self.phone = phone
    
    def is_seller(self) -> bool:
        """Check if user is a seller."""
        return self.role == UserRole.SELLER
    
    def is_buyer(self) -> bool:
        """Check if user is a buyer."""
        return self.role == UserRole.BUYER
    
    def is_admin(self) -> bool:
        """Check if user is an admin."""
        return self.role == UserRole.ADMIN
