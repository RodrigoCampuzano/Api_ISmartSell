from typing import Optional, List
from sqlalchemy import select, or_, func
from sqlalchemy.ext.asyncio import AsyncSession
from src.modules.stores.domain.entities.store import Store
from src.modules.stores.domain.entities.delivery_point import DeliveryPoint
from src.modules.stores.domain.repositories.store_repository import StoreRepository
from src.modules.stores.infrastructure.models.store_model import StoreModel, DeliveryPointModel


class StoreRepositoryImpl(StoreRepository):
    """Implementation of StoreRepository using SQLAlchemy."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    def _to_entity(self, model: StoreModel) -> Store:
        """Convert ORM model to domain entity."""
        return Store(
            id=model.id,
            seller_id=model.seller_id,
            name=model.name,
            slug=model.slug,
            description=model.description,
            address=model.address,
            lat=float(model.lat) if model.lat else None,
            lng=float(model.lng) if model.lng else None
        )
    
    def _to_model(self, entity: Store) -> StoreModel:
        """Convert domain entity to ORM model."""
        return StoreModel(
            id=entity.id,
            seller_id=entity.seller_id,
            name=entity.name,
            slug=entity.slug,
            description=entity.description,
            address=entity.address,
            lat=entity.lat,
            lng=entity.lng
        )
    
    def _point_to_entity(self, model: DeliveryPointModel) -> DeliveryPoint:
        """Convert delivery point model to entity."""
        return DeliveryPoint(
            id=model.id,
            store_id=model.store_id,
            name=model.name,
            address=model.address,
            lat=float(model.lat) if model.lat else None,
            lng=float(model.lng) if model.lng else None
        )
    
    async def create(self, entity: Store) -> Store:
        """Create a new store."""
        model = self._to_model(entity)
        self.db.add(model)
        await self.db.flush()
        await self.db.refresh(model)
        return self._to_entity(model)
    
    async def get_by_id(self, id: int) -> Optional[Store]:
        """Get store by ID."""
        result = await self.db.execute(select(StoreModel).where(StoreModel.id == id))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_slug(self, slug: str) -> Optional[Store]:
        """Get store by slug."""
        result = await self.db.execute(select(StoreModel).where(StoreModel.slug == slug))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_seller(self, seller_id: int) -> List[Store]:
        """Get all stores for a seller."""
        result = await self.db.execute(select(StoreModel).where(StoreModel.seller_id == seller_id))
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def search(self, query: str, skip: int = 0, limit: int = 100) -> List[Store]:
        """Search stores by name or description."""
        result = await self.db.execute(
            select(StoreModel)
            .where(
                or_(
                    StoreModel.name.ilike(f"%{query}%"),
                    StoreModel.description.ilike(f"%{query}%")
                )
            )
            .offset(skip)
            .limit(limit)
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def get_nearby(self, lat: float, lng: float, radius_km: float = 10.0) -> List[Store]:
        """Get stores within radius using Haversine formula."""
        # Simplified distance calculation (for production, use PostGIS)
        # Distance in km using Haversine approximation
        distance_formula = func.sqrt(
            func.pow((StoreModel.lat - lat) * 111.32, 2) +
            func.pow((StoreModel.lng - lng) * 111.32 * func.cos(func.radians(lat)), 2)
        )
        
        result = await self.db.execute(
            select(StoreModel)
            .where(
                StoreModel.lat.isnot(None),
                StoreModel.lng.isnot(None)
            )
            .having(distance_formula <= radius_km)
            .order_by(distance_formula)
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def update(self, entity: Store) -> Store:
        """Update existing store."""
        result = await self.db.execute(select(StoreModel).where(StoreModel.id == entity.id))
        model = result.scalar_one_or_none()
        
        if model:
            model.name = entity.name
            model.slug = entity.slug
            model.description = entity.description
            model.address = entity.address
            model.lat = entity.lat
            model.lng = entity.lng
            await self.db.flush()
            await self.db.refresh(model)
            return self._to_entity(model)
        
        return None
    
    async def delete(self, id: int) -> bool:
        """Delete store by ID."""
        result = await self.db.execute(select(StoreModel).where(StoreModel.id == id))
        model = result.scalar_one_or_none()
        
        if model:
            await self.db.delete(model)
            await self.db.flush()
            return True
        
        return False
    
    async def list(self, skip: int = 0, limit: int = 100) -> List[Store]:
        """List stores with pagination."""
        result = await self.db.execute(select(StoreModel).offset(skip).limit(limit))
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def add_delivery_point(self, point: DeliveryPoint) -> DeliveryPoint:
        """Add delivery point to store."""
        model = DeliveryPointModel(
            store_id=point.store_id,
            name=point.name,
            address=point.address,
            lat=point.lat,
            lng=point.lng
        )
        self.db.add(model)
        await self.db.flush()
        await self.db.refresh(model)
        return self._point_to_entity(model)
    
    async def get_delivery_points(self, store_id: int) -> List[DeliveryPoint]:
        """Get all delivery points for a store."""
        result = await self.db.execute(
            select(DeliveryPointModel).where(DeliveryPointModel.store_id == store_id)
        )
        models = result.scalars().all()
        return [self._point_to_entity(model) for model in models]
