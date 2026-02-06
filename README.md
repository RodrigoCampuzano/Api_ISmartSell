# ISmartSell API

Complete REST API for an e-commerce platform built with **Python**, **FastAPI**, and **Hexagonal Architecture**.

## Features

- ✅ **Authentication**: JWT-based user registration and login with role-based access (Buyer, Seller, Admin)
- ✅ **Stores**: Seller store management with delivery points and geospatial nearby search
- ✅ **Products**: Product catalog with stock management and seller authorization
- ✅ **Orders**: Order processing with state machine (PENDING → RESERVED → PAID → READY → DELIVERED)
- ✅ **Reservation System**: Automatic timeout and cancellation for unpaid reservations with stock restoration
- ✅ **Payments**: Payment processing with 1% platform commission tracking
- ✅ **QR Validation**: QR code-based order pickup validation
- ✅ **Background Tasks**: Automatic cancellation of expired reservations every 5 minutes
- ✅ **Swagger Documentation**: Interactive API documentation at `/docs`

## Architecture

### Hexagonal Architecture (Ports & Adapters)
- **Domain Layer**: Business entities, value objects, repository interfaces (ports)
- **Application Layer**: Use cases orchestrating business logic
- **Infrastructure Layer**: Repository implementations, ORM models, external services
- **Presentation Layer**: FastAPI routes and DTOs

### Vertical Slicing
Each module is self-contained with all layers:
```
modules/auth/
├── domain/           # Entities, repository interfaces
├── application/      # Use cases, DTOs
├── infrastructure/   # Repository implementations, ORM models
└── presentation/     # FastAPI routes
```

## Installation

### Prerequisites
- Python 3.10+
- PostgreSQL or MySQL

### Setup

1. Clone the repository
```bash
cd /home/rodrigo/Documentos/Api_ISamrtSell
```

2. Create virtual environment
```bash
python -m venv venv
source venv/bin/activate  # On Linux/Mac
# or
venv\Scripts\activate  # On Windows
```

3. Install dependencies
```bash
pip install -r requirements.txt
```

4. Configure environment variables
```bash
cp .env.example .env
# Edit .env with your database credentials and settings
```

5. Run database migrations
```bash
alembic upgrade head
```

6. Start the server
```bash
uvicorn src.main:app --reload
```

The API will be available at `http://localhost:8000`

## API Documentation

- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login and get JWT token
- `GET /api/auth/me` - Get current user profile

### Stores
- `POST /api/stores` - Create store (seller only)
- `GET /api/stores` - List stores (supports search and nearby)
- `GET /api/stores/{id}` - Get store details
- `POST /api/stores/{id}/points` - Add delivery point
- `GET /api/stores/{id}/points` - Get delivery points

### Products
- `POST /api/stores/{storeId}/products` - Create product (seller only)
- `GET /api/stores/{storeId}/products` - List products by store
- `GET /api/products/{id}` - Get product details
- `PUT /api/products/{id}` - Update product (seller only)
- `DELETE /api/products/{id}` - Delete product (seller only)

### Orders
- `POST /api/orders` - Create order
- `GET /api/orders/{id}` - Get order details
- `GET /api/users/{id}/orders` - Get user order history
- `PUT /api/orders/{id}/cancel` - Cancel order
- `PUT /api/orders/{id}/mark-ready` - Mark order ready (seller only)

### Payments
- `POST /api/payments/{orderId}/pay` - Initiate payment
- `POST /api/payments/{paymentId}/complete` - Complete payment (demo endpoint)
- `POST /api/payments/webhook` - Payment provider webhook
- `GET /api/payments/{id}` - Get payment details

### QR Validation
- `POST /api/qr/validate` - Validate QR code and deliver order

## Usage Example

### 1. Register as Seller
```bash
curl -X POST "http://localhost:8000/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seller@example.com",
    "password": "password123",
    "role": "SELLER",
    "full_name": "John Seller"
  }'
```

### 2. Create Store
```bash
curl -X POST "http://localhost:8000/api/stores" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mi Tienda",
    "slug": "mi-tienda",
    "description": "Fresh products",
    "lat": 19.4326,
    "lng": -99.1332
  }'
```

### 3. Add Product
```bash
curl -X POST "http://localhost:8000/api/stores/1/products" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tomate",
    "price": 12.50,
    "stock": 100
  }'
```

### 4. Create Order (as Buyer)
```bash
curl -X POST "http://localhost:8000/api/orders" \
  -H "Authorization: Bearer BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": 1,
    "items": [{"product_id": 1, "quantity": 5}],
    "payment_method": "ONLINE"
  }'
```

### 5. Complete Payment
```bash
curl -X POST "http://localhost:8000/api/payments/1/complete" \
  -H "Content-Type: application/json"
```

### 6. Validate QR for Pickup
```bash
curl -X POST "http://localhost:8000/api/qr/validate" \
  -H "Content-Type: application/json" \
  -d '{
    "qr_token": "QR_TOKEN_FROM_ORDER"
  }'
```

## Development

### Run tests
```bash
pytest tests/ -v
```

### Create new migration
```bash
alembic revision --autogenerate -m "description"
```

### Apply migrations
```bash
alembic upgrade head
```

## Project Structure

```
Api_ISamrtSell/
├── src/
│   ├── main.py                 # FastAPI application entry point
│   ├── config.py               # Settings and configuration
│   ├── core/                   # Shared core components
│   │   ├── domain/             # Base entities and exceptions
│   │   └── infrastructure/     # Database, security, scheduler
│   ├── shared/                 # Shared utilities
│   └── modules/                # Feature modules (vertical slices)
│       ├── auth/
│       ├── stores/
│       ├── products/
│       ├── orders/
│       ├── payments/
│       └── qr/
├── alembic/                    # Database migrations
├── tests/
├── requirements.txt
├── .env
└── README.md
```

## License

MIT
