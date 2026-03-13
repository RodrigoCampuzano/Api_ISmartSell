# POS App — API en Go (Arquitectura Hexagonal)

## Estructura del proyecto

```
pos-api/
├── cmd/api/main.go                          ← Entry point + inyección de dependencias
├── schema.sql                               ← Schema MySQL
│
├── internal/
│   ├── domain/                              ← Núcleo del negocio (sin dependencias externas)
│   │   ├── user/       entity.go · repository.go
│   │   ├── business/   entity.go · repository.go
│   │   ├── product/    entity.go · repository.go
│   │   └── order/      entity.go · repository.go
│   │
│   ├── application/services/               ← Casos de uso (puertos de entrada)
│   │   ├── user_service.go
│   │   ├── business_service.go
│   │   ├── product_service.go
│   │   └── order_service.go
│   │
│   └── infrastructure/                     ← Adaptadores
│       ├── persistence/mysql/              ← Adaptadores de salida (BD)
│       │   ├── user_repo.go
│       │   ├── business_repo.go
│       │   ├── product_repo.go
│       │   └── order_repo.go
│       └── http/                           ← Adaptadores de entrada (REST)
│           ├── router.go
│           ├── middleware/auth.go
│           └── handler/
│               ├── user_handler.go
│               ├── business_handler.go
│               ├── product_handler.go
│               └── order_handler.go
│
└── pkg/
    ├── config/config.go
    ├── jwt/jwt.go
    ├── qr/qr.go
    └── response/response.go
```

## Capas (Hexagonal)

```
┌───────────────────────────────────────────┐
│          Adaptadores de entrada           │
│       HTTP Handlers (chi router)          │
├───────────────────────────────────────────┤
│         Puertos de entrada (interfaces)   │
│          Application Services            │
├───────────────────────────────────────────┤
│             DOMINIO (núcleo)              │
│    Entities · Domain Errors · Logic       │
├───────────────────────────────────────────┤
│         Puertos de salida (interfaces)    │
│          Repository interfaces            │
├───────────────────────────────────────────┤
│         Adaptadores de salida             │
│          MySQL Repositories               │
└───────────────────────────────────────────┘
```

## Endpoints

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/api/v1/auth/register` | público | Registrar usuario |
| POST | `/api/v1/auth/login` | público | Login → JWT |
| GET | `/api/v1/users/me` | auth | Perfil propio |
| GET | `/api/v1/businesses?lat=&lng=&radius=` | auth | Negocios cercanos |
| GET | `/api/v1/businesses/{id}` | auth | Detalle + puntos de entrega |
| POST | `/api/v1/businesses` | seller | Crear negocio |
| GET | `/api/v1/businesses/mine` | seller | Mis negocios |
| POST | `/api/v1/businesses/{id}/delivery-points` | seller | Agregar punto de entrega |
| GET | `/api/v1/businesses/{businessId}/products` | auth | Productos del negocio |
| POST | `/api/v1/businesses/{businessId}/products` | seller | Crear producto |
| GET | `/api/v1/products/{id}` | auth | Detalle producto |
| PUT | `/api/v1/products/{id}` | seller | Editar producto |
| DELETE | `/api/v1/products/{id}` | seller | Eliminar producto |
| POST | `/api/v1/orders` | buyer | Crear orden (online/apartado) |
| GET | `/api/v1/orders/my` | buyer | Mis órdenes |
| POST | `/api/v1/orders/{id}/cancel` | buyer | Cancelar orden |
| GET | `/api/v1/orders/{id}` | auth | Ver orden |
| GET | `/api/v1/businesses/{businessId}/orders` | seller | Órdenes del negocio |
| POST | `/api/v1/orders/scan` | seller | Escanear QR → entregar |

## Levantar el proyecto

```bash
# 1. Variables de entorno
export DSN="user:pass@tcp(localhost:3306)/pos_app?parseTime=true&charset=utf8mb4"
export JWT_SECRET="mi-secreto-seguro"
export PORT=8080

# 2. Crear la BD
mysql -u root -p < schema.sql

# 3. Dependencias
go mod tidy

# 4. Ejecutar
go run ./cmd/api/main.go
```

## Modelo de comisión

Cada orden guarda `commission = total * 0.01` (1%) en la tabla `payments`.
El job de background cancela automáticamente apartados expirados cada 5 minutos.
