from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List

from src.core.infrastructure.database import get_db
from src.core.infrastructure.security import get_current_user
from src.modules.products.application.dtos import ProductRequest, ProductResponse
from src.modules.products.domain.entities.product import Product
from src.modules.products.infrastructure.repositories.product_repository_impl import ProductRepositoryImpl
from src.modules.stores.infrastructure.repositories.store_repository_impl import StoreRepositoryImpl

router = APIRouter(tags=["Products"])


@router.post("/stores/{store_id}/products", response_model=ProductResponse, status_code=status.HTTP_201_CREATED)
async def create_product(
    store_id: int,
    request: ProductRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Create a new product (Seller only).
    
    Only the store owner can add products.
    """
    store_repo = StoreRepositoryImpl(db)
    store = await store_repo.get_by_id(store_id)
    
    if not store:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Store {store_id} not found"
        )
    
    if store.seller_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only add products to your own stores"
        )
    
    product_repo = ProductRepositoryImpl(db)
    product = Product(
        store_id=store_id,
        name=request.name,
        sku=request.sku,
        description=request.description,
        price=request.price,
        stock=request.stock,
        image_url=request.image_url,
        active=request.active
    )
    
    created = await product_repo.create(product)
    return ProductResponse.model_validate(created)


@router.get("/stores/{store_id}/products", response_model=List[ProductResponse])
async def get_store_products(
    store_id: int,
    db: AsyncSession = Depends(get_db)
):
    """Get all products for a store."""
    product_repo = ProductRepositoryImpl(db)
    products = await product_repo.get_by_store(store_id)
    return [ProductResponse.model_validate(p) for p in products]


@router.get("/products/{product_id}", response_model=ProductResponse)
async def get_product(
    product_id: int,
    db: AsyncSession = Depends(get_db)
):
    """Get product by ID."""
    product_repo = ProductRepositoryImpl(db)
    product = await product_repo.get_by_id(product_id)
    
    if not product:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product {product_id} not found"
        )
    
    return ProductResponse.model_validate(product)


@router.put("/products/{product_id}", response_model=ProductResponse)
async def update_product(
    product_id: int,
    request: ProductRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Update product (Seller only).
    
    Only the store owner can update products.
    """
    product_repo = ProductRepositoryImpl(db)
    product = await product_repo.get_by_id(product_id)
    
    if not product:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product {product_id} not found"
        )
    
    # Check ownership
    store_repo = StoreRepositoryImpl(db)
    store = await store_repo.get_by_id(product.store_id)
    
    if store.seller_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only update your own products"
        )
    
    # Update product
    product.name = request.name
    product.sku = request.sku
    product.description = request.description
    product.price = request.price
    product.stock = request.stock
    product.image_url = request.image_url
    product.active = request.active
    
    updated = await product_repo.update(product)
    return ProductResponse.model_validate(updated)


@router.delete("/products/{product_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_product(
    product_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Delete product (Seller only).
    
    Only the store owner can delete products.
    """
    product_repo = ProductRepositoryImpl(db)
    product = await product_repo.get_by_id(product_id)
    
    if not product:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product {product_id} not found"
        )
    
    # Check ownership
    store_repo = StoreRepositoryImpl(db)
    store = await store_repo.get_by_id(product.store_id)
    
    if store.seller_id != current_user.id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You can only delete your own products"
        )
    
    await product_repo.delete(product_id)
