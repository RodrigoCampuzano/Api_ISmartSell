from src.modules.auth.domain.entities.user import User
from src.modules.auth.domain.repositories.user_repository import UserRepository
from src.core.domain.exceptions import ValidationException, UnauthorizedException
from src.core.infrastructure.security import verify_password


class LoginUseCase:
    """Use case for user login."""
    
    def __init__(self, user_repository: UserRepository):
        self.user_repository = user_repository
    
    async def execute(self, email: str, password: str) -> User:
        """Authenticate user with email and password."""
        
        # Get user by email
        user = await self.user_repository.get_by_email(email)
        
        if not user:
            raise UnauthorizedException("Invalid email or password")
        
        # Verify password
        if not verify_password(password, user.password_hash):
            raise UnauthorizedException("Invalid email or password")
        
        return user
