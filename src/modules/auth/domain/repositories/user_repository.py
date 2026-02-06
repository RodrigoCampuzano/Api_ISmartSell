from abc import abstractmethod
from typing import Optional
from src.core.domain.base_repository import BaseRepository
from src.modules.auth.domain.entities.user import User


class UserRepository(BaseRepository[User]):
    """User repository interface (port)."""
    
    @abstractmethod
    async def get_by_email(self, email: str) -> Optional[User]:
        """Get user by email address."""
        pass
    
    @abstractmethod
    async def get_by_username(self, username: str) -> Optional[User]:
        """Get user by username."""
        pass
    
    @abstractmethod
    async def email_exists(self, email: str) -> bool:
        """Check if email already exists."""
        pass
