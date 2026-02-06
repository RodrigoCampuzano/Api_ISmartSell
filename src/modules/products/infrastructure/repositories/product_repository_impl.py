from typing import Optional, List
from sqlalchemy import select, or_
from sqlalchemy.ext.asyncio import AsyncSession
from src.modules.products.domain.entities.product import Product
from src.modules.products.domain.repositories.product_repository import ProductRepository
from src.modules.products.infrastructure.models.product_model import ProductModel


class ProductRepositoryImpl(ProductRepository):
    """Implementation of ProductRepository using SQLAlchemy."""
    
    def __init__(self, db: AsyncSession):
        self.db = db
    
    def _to_entity(self, model: ProductModel) -> Product:
        """Convert ORM model to domain entity."""
        return Product(
            id=model.id,
            store_id=model.store_id,
            name=model.name,
            sku=model.sku,
            description=model.description,
            price=float(model.price),
            stock=model.stock,
            image_url=model.image_url,
            active=model.active
        )
    
    def _to_model(self, entity: Product) -> ProductModel:
        """Convert domain entity to ORM model."""
        return ProductModel(
            id=entity.id,
            store_id=entity.store_id,
            name=entity.name,
            sku=entity.sku,
            description=entity.description,
            price=entity.price,
            stock=entity.stock,
            image_url=entity.image_url,
            active=entity.active
        )
    
    async def create(self, entity: Product) -> Product:
        """Create a new product."""
        model = self._to_model(entity)
        self.db.add(model)
        await self.db.flush()
        await self.db.refresh(model)
        return self._to_entity(model)
    
    async def get_by_id(self, id: int) -> Optional[Product]:
        """Get product by ID."""
        result = await self.db.execute(select(ProductModel).where(ProductModel.id == id))
        model = result.scalar_one_or_none()
        return self._to_entity(model) if model else None
    
    async def get_by_store(self, store_id: int, skip: int = 0, limit: int = 100) -> List[Product]:
        """Get all products for a store."""
        result = await self.db.execute(
            select(ProductModel)
            .where(ProductModel.store_id == store_id)
            .offset(skip)
            .limit(limit)
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def search(self, query: str, skip: int = 0, limit: int = 100) -> List[Product]:
        """Search products by name or description."""
        result = await self.db.execute(
            select(ProductModel)
            .where(
                or_(
                    ProductModel.name.ilike(f"%{query}%"),
                    ProductModel.description.ilike(f"%{query}%")
                )
            )
            .offset(skip)
            .limit(limit)
        )
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
    
    async def update(self, entity: Product) -> Product:
        """Update existing product."""
        result = await self.db.execute(select(ProductModel).where(ProductModel.id == entity.id))
        model = result.scalar_one_or_none()
        
        if model:
            model.name = entity.name
            model.sku = entity.sku
            model.description = entity.description
            model.price = entity.price
            model.stock = entity.stock
            model.image_url = entity.image_url
            model.active = entity.active
            await self.db.flush()
            await self.db.refresh(model)
            return self._to_entity(model)
        
        return None
    
    async def delete(self, id: int) -> bool:
        """Delete product by ID."""
        result = await self.db.execute(select(ProductModel).where(ProductModel.id == id))
        model = result.scalar_one_or_none()
        
        if model:
            await self.db.delete(model)
            await self.db.flush()
            return True
        
        return False
    
    async def list(self, skip: int = 0, limit: int = 100) -> List[Product]:
        """List products with pagination."""
        result = await self.db.execute(select(ProductModel).offset(skip).limit(limit))
        models = result.scalars().all()
        return [self._to_entity(model) for model in models]
