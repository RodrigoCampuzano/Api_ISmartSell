from src.modules.auth.domain.entities.user import User, UserRole
from src.modules.auth.domain.repositories.user_repository import UserRepository
from src.core.domain.exceptions import ValidationException
from src.core.infrastructure.security import hash_password


class RegisterUseCase:
    """Use case for user registration."""
    
    def __init__(self, user_repository: UserRepository):
        self.user_repository = user_repository
    
    async def execute(
        self,
        email: str,
        password: str,
        role: UserRole = UserRole.BUYER,
        username: str = None,
        full_name: str = None,
        phone: str = None
    ) -> User:
        """Register a new user."""
        
        # Check if email already exists
        if await self.user_repository.email_exists(email):
            raise ValidationException(f"Email {email} is already registered")
        
        # Check if username exists (if provided)
        if username:
            existing_user = await self.user_repository.get_by_username(username)
            if existing_user:
                raise ValidationException(f"Username {username} is already taken")
        
        # Hash password
        password_hash = hash_password(password)
        
        # Create user entity
        user = User(
            email=email,
            password_hash=password_hash,
            role=role,
            username=username,
            full_name=full_name,
            phone=phone
        )
        
        # Persist user
        return await self.user_repository.create(user)
