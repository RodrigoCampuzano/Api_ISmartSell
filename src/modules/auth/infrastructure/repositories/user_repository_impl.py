from typing import Optional, List
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from src.modules.auth.domain.entities.user import User, UserRole
from src.modules.auth.domain.repositories.user_repository import UserRepository
from src.modules.auth.infrastructure.models.user_model import UserModel, UserRoleEnum


class UserRepositoryImpl(UserRepository):
    """Implementation of UserRepository using SQLAlchemy."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    def _to_entity(self, model: UserModel) -> User:
        """Convert ORM model to domain entity."""
        return User(
            id=model.id,
            email=model.email,
            username=model.username,
            password_hash=model.password_hash,
            role=UserRole(model.role.value),
            full_name=model.full_name,
            phone=model.phone
        )
    
    def _to_model(self, entity: User) -> UserModel:
        """Convert domain entity to ORM model."""
        return UserModel(
            id=entity.id,
            email=entity.email,
            username=entity.username,
            password_hash=entity.password_hash,
            role=UserRoleEnum(entity.role.value),
            full_name=entity.full_name,
            phone=entity.phone
        )
    
    async def create(self, entity: User) -> User:
        """Create a new user."""
        model = self._to_model(entity)
        self.db.add(model)
        await self.db.flush()
        await self.db.refresh(model)
        return self._to_entity(model)
    
    async def get_by_id(self, id: int) -> Optional[User]:
        """Get user by ID."""
        result = await self.db.execute(select(UserModel).where(UserModel.id == id))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_email(self, email: str) -> Optional[User]:
        """Get user by email."""
        result = await self.db.execute(select(UserModel).where(UserModel.email == email))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_username(self, username: str) -> Optional[User]:
        """Get user by username."""
        result = await self.db.execute(select(UserModel).where(UserModel.username == username))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def email_exists(self, email: str) -> bool:
        """Check if email exists."""
        result = await self.db.execute(select(UserModel.id).where(UserModel.email == email))
        return result.scalar_one_or_none() is not None
    
    async def update(self, entity: User) -> User:
        """Update existing user."""
        result = await self.db.execute(select(UserModel).where(UserModel.id == entity.id))
        model = result.scalar_one_or_none()
        
        if model:
            model.email = entity.email
            model.username = entity.username
            model.full_name = entity.full_name
            model.phone = entity.phone
            model.role = UserRoleEnum(entity.role.value)
            await self.db.flush()
            await self.db.refresh(model)
            return self._to_entity(model)
        
        return None
    
    async def delete(self, id: int) -> bool:
        """Delete user by ID."""
        result = await self.db.execute(select(UserModel).where(UserModel.id == id))
        model = result.scalar_one_or_none()
        
        if model:
            await self.db.delete(model)
            await self.db.flush()
            return True
        
        return False
    
    async def list(self, skip: int = 0, limit: int = 100) -> List[User]:
        """List users with pagination."""
        result = await self.db.execute(select(UserModel).offset(skip).limit(limit))
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
