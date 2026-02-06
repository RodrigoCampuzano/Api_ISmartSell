from fastapi import FastAPI, HTTPException, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager
import logging

from src.config import settings
from src.core.infrastructure.database import engine, Base
from src.core.infrastructure.scheduler import start_scheduler
from src.core.domain.exceptions import (
    DomainException,
    EntityNotFoundException,
    ValidationException,
    UnauthorizedException,
    BusinessRuleViolationException
)

# Import all routes
from src.modules.auth.presentation.routes import router as auth_router
from src.modules.stores.presentation.routes import router as stores_router
from src.modules.products.presentation.routes import router as products_router
from src.modules.orders.presentation.routes import router as orders_router
from src.modules.payments.presentation.routes import router as payments_router
from src.modules.qr.presentation.routes import router as qr_router

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


# Lifespan context manager for startup/shutdown events
@asynccontextmanager
async def lifespan(app: FastAPI):
    """Handle startup and shutdown events."""
    # Startup
    logger.info("Starting ISmartSell API...")
    
    # Create database tables
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    logger.info("Database tables created")
    
    # Start background scheduler
    scheduler = start_scheduler()
    
    yield
    
    # Shutdown
    logger.info("Shutting down ISmartSell API...")
    scheduler.shutdown()


# Create FastAPI app
app = FastAPI(
    title=settings.PROJECT_NAME,
    version=settings.VERSION,
    description="""
    ## ISmartSell E-Commerce API
    
    Complete REST API with:
    - **Authentication**: JWT-based user registration and login
    - **Stores**: Seller store management with delivery points and geospatial search
    - **Products**: Product catalog with stock management
    - **Orders**: Order processing with reservation timeouts and state machine
    - **Payments**: Payment processing with 1% platform commission
    - **QR Validation**: QR code-based order pickup validation
    
    Built with **Hexagonal Architecture**, **Clean Architecture**, and **Vertical Slicing**.
    """,
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    swagger_ui_parameters={
        "persistAuthorization": True,
    }
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Exception handlers
@app.exception_handler(EntityNotFoundException)
async def entity_not_found_handler(request: Request, exc: EntityNotFoundException):
    return JSONResponse(
        status_code=status.HTTP_404_NOT_FOUND,
        content={"detail": str(exc)}
    )


@app.exception_handler(ValidationException)
async def validation_exception_handler(request: Request, exc: ValidationException):
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": str(exc)}
    )


@app.exception_handler(UnauthorizedException)
async def unauthorized_exception_handler(request: Request, exc: UnauthorizedException):
    return JSONResponse(
        status_code=status.HTTP_401_UNAUTHORIZED,
        content={"detail": str(exc)}
    )


@app.exception_handler(BusinessRuleViolationException)
async def business_rule_violation_handler(request: Request, exc: BusinessRuleViolationException):
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": str(exc)}
    )


# Register routers
app.include_router(auth_router, prefix=settings.API_V1_PREFIX)
app.include_router(stores_router, prefix=settings.API_V1_PREFIX)
app.include_router(products_router, prefix=settings.API_V1_PREFIX)
app.include_router(orders_router, prefix=settings.API_V1_PREFIX)
app.include_router(payments_router, prefix=settings.API_V1_PREFIX)
app.include_router(qr_router, prefix=settings.API_V1_PREFIX)


@app.get("/")
async def root():
    """API root endpoint."""
    return {
        "message": "Welcome to ISmartSell API",
        "version": settings.VERSION,
        "docs": "/docs",
        "health": "/health"
    }


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "version": settings.VERSION
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=True
    )
