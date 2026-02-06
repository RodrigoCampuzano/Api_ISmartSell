from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.ext.asyncio import AsyncSession
from typing import List, Optional

from src.core.infrastructure.database import get_db
from src.core.infrastructure.security import get_current_user
from src.modules.stores.application.dtos import (
    StoreRequest,
    StoreResponse,
    DeliveryPointRequest,
    DeliveryPointResponse
)
from src.modules.stores.domain.entities.store import Store
from src.modules.stores.domain.entities.delivery_point import DeliveryPoint
from src.modules.stores.infrastructure.repositories.store_repository_impl import StoreRepositoryImpl
from src.modules.auth.domain.entities.user import UserRole

router = APIRouter(prefix="/stores", tags=["Stores"])


@router.post("", response_model=StoreResponse, status_code=status.HTTP_201_CREATED)
async def create_store(
    request: StoreRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Create a new store (Seller only).
    
    Requires SELLER role.
    """
    if current_user.role != UserRole.SELLER:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Only sellers can create stores"
        )
    
    store_repo = StoreRepositoryImpl(db)
    
    # Check if slug exists
    existing = await store_repo.get_by_slug(request.slug)
    if existing:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Store with slug '{request.slug}' already exists"
        )
    
    store = Store(
        seller_id=current_user.id,
        name=request.name,
        slug=request.slug,
        description=request.description,
        address=request.address,
        lat=request.lat,
        lng=request.lng
    )
    
    created = await store_repo.create(store)
    return StoreResponse.model_validate(created)


@router.get("", response_model=List[StoreResponse])
async def get_stores(
    q: Optional[str] = Query(None, description="Search query"),
    lat: Optional[float] = Query(None, description="Latitude for nearby search"),
    lng: Optional[float] = Query(None, description="Longitude for nearby search"),
    radius: float = Query(10.0, description="Search radius in km"),
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    db: AsyncSession = Depends(get_db)
):
    """
    List stores with optional filters.
    
    - **q**: Search by name or description
    - **lat/lng/radius**: Search nearby stores within radius (km)
    """
    store_repo = StoreRepositoryImpl(db)
    
    if lat is not None and lng is not None:
        stores = await store_repo.get_nearby(lat, lng, radius)
    elif q:
        stores = await store_repo.search(q, skip, limit)
    else:
        stores = await store_repo.list(skip, limit)
    
    return [StoreResponse.model_validate(s) for s in stores]


@router.get("/{store_id}", response_model=StoreResponse)
async def get_store(
    store_id: int,
    db: AsyncSession = Depends(get_db)
):
    """Get store by ID."""
    store_repo = StoreRepositoryImpl(db)
    store = await store_repo.get_by_id(store_id)
    
    if not store:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Store {store_id} not found"
        )
    
    return StoreResponse.model_validate(store)


@router.post("/{store_id}/points", response_model=DeliveryPointResponse, status_code=status.HTTP_201_CREATED)
async def add_delivery_point(
    store_id: int,
    request: DeliveryPointRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Add delivery point to store (Seller only).
    
    Only the store owner can add delivery points.
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
            detail="You can only add delivery points to your own stores"
        )
    
    point = DeliveryPoint(
        store_id=store_id,
        name=request.name,
        address=request.address,
        lat=request.lat,
        lng=request.lng
    )
    
    created = await store_repo.add_delivery_point(point)
    return DeliveryPointResponse.model_validate(created)


@router.get("/{store_id}/points", response_model=List[DeliveryPointResponse])
async def get_delivery_points(
    store_id: int,
    db: AsyncSession = Depends(get_db)
):
    """Get all delivery points for a store."""
    store_repo = StoreRepositoryImpl(db)
    points = await store_repo.get_delivery_points(store_id)
    return [DeliveryPointResponse.model_validate(p) for p in points]


@router.put("/{store_id}", response_model=StoreResponse)
async def update_store(
    store_id: int,
    request: StoreRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Update store (Seller only).
    
    Only the store owner can update it.
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
            detail="You can only update your own stores"
        )
    
    # Check if new slug already exists (if it's different)
    if request.slug != store.slug:
        existing = await store_repo.get_by_slug(request.slug)
        if existing:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Store with slug '{request.slug}' already exists"
            )
    
    # Update fields
    store.name = request.name
    store.slug = request.slug
    store.description = request.description
    store.address = request.address
    store.lat = request.lat
    store.lng = request.lng
    
    updated = await store_repo.update(store)
    return StoreResponse.model_validate(updated)


@router.delete("/{store_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_store(
    store_id: int,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    Delete store (Seller only).
    
    Only the store owner can delete it.
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
            detail="You can only delete your own stores"
        )
    
    await store_repo.delete(store_id)
    return None

