# 🛒 iSmartSell — POS API

API REST para un sistema de Punto de Venta (POS) con soporte geoespacial. Permite a vendedores gestionar negocios y productos, y a compradores realizar pedidos en línea o reservar para recoger.

**Base URL:** `https://apismart.serviciocdn.icu/api/v1`

## 🏗️ Stack Tecnológico

| Componente | Tecnología |
|---|---|
| Lenguaje | Go 1.22 |
| Router HTTP | chi v5 |
| Base de datos | PostgreSQL 16 + PostGIS |
| Autenticación | JWT (Bearer token) |
| ORM / Driver | sqlx + lib/pq |
| Hashing | bcrypt |

## 📁 Estructura del Proyecto

```
Api_ISmartSell/
├── main.go                                  ← Entry point + inyección de dependencias
├── schema.sql                               ← Schema PostgreSQL + PostGIS
├── deploy.sh                                ← Script de deploy completo
├── go.mod
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
│       ├── persistence/postgres/           ← Adaptadores de salida (BD)
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

## 🔐 Autenticación

Todas las rutas protegidas requieren el header:

```
Authorization: Bearer <token>
```

El token se obtiene al registrarse o iniciar sesión. Contiene `user_id` y `role`. Expira en 72 horas (configurable con `JWT_TTL_HOURS`).

## ⚙️ Variables de Entorno

| Variable | Default | Descripción |
|---|---|---|
| `PORT` | `8080` | Puerto del servidor |
| `DSN` | `postgres://postgres:postgres@localhost:5432/pos_app?sslmode=disable` | Conexión PostgreSQL |
| `JWT_SECRET` | `change-me-in-production` | Clave secreta para firmar JWT |
| `JWT_TTL_HOURS` | `72` | Horas de vida del token |

## 📌 Formato de Respuestas

Todas las respuestas exitosas siguen el formato:

```json
{ "data": { ... } }
```

Los errores:

```json
{ "error": "mensaje descriptivo" }
```

---

## 📖 Endpoints

### 🔓 Auth (Público)

---

#### `POST /api/v1/auth/register`

Registra un nuevo usuario y devuelve un token JWT.

**Request:**
```json
{
  "name": "Juan Pérez",
  "email": "juan@email.com",
  "password": "miPassword123",
  "role": "seller"
}
```
> `role`: `"seller"` o `"buyer"`

**Response (`201`):**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "Juan Pérez",
      "email": "juan@email.com",
      "password": "$2a$10$...",
      "role": "seller",
      "active": true,
      "created_at": "2026-03-22T20:00:00Z",
      "updated_at": "2026-03-22T20:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Errores:**
| Código | Descripción |
|---|---|
| `400` | Body inválido o rol inválido |
| `409` | Email ya registrado |

---

#### `POST /api/v1/auth/login`

Inicia sesión y devuelve un token JWT.

**Request:**
```json
{
  "email": "juan@email.com",
  "password": "miPassword123"
}
```

**Response (`200`):**
```json
{
  "data": {
    "user": { ... },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Errores:**
| Código | Descripción |
|---|---|
| `400` | Body inválido |
| `401` | Credenciales incorrectas |

---

### 👤 Usuarios (Autenticado)

---

#### `GET /api/v1/users/me`

Devuelve el perfil del usuario autenticado.

**Headers:** `Authorization: Bearer <token>`

**Response (`200`):**
```json
{
  "data": {
    "id": "uuid",
    "name": "Juan Pérez",
    "email": "juan@email.com",
    "password": "",
    "role": "seller",
    "active": true,
    "created_at": "2026-03-22T20:00:00Z",
    "updated_at": "2026-03-22T20:00:00Z"
  }
}
```

---

### 🏪 Negocios

---

#### `POST /api/v1/businesses` 🔒 seller

Crea un nuevo negocio.

**Request:**
```json
{
  "name": "Mi Tienda",
  "description": "Tienda de abarrotes",
  "type": "grocery",
  "latitude": 20.6597,
  "longitude": -103.3496
}
```

**Response (`201`):**
```json
{
  "data": {
    "id": "uuid",
    "owner_id": "uuid",
    "name": "Mi Tienda",
    "description": "Tienda de abarrotes",
    "type": "grocery",
    "latitude": 20.6597,
    "longitude": -103.3496,
    "active": true,
    "created_at": "2026-03-22T20:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z"
  }
}
```

---

#### `GET /api/v1/businesses?lat=20.65&lng=-103.34&radius=5` 🔒 autenticado

Lista negocios cercanos a una ubicación. Usa **PostGIS** (`ST_DWithin`) para búsquedas geoespaciales eficientes.

**Query Params:**
| Param | Tipo | Default | Descripción |
|---|---|---|---|
| `lat` | float | requerido | Latitud del punto de búsqueda |
| `lng` | float | requerido | Longitud del punto de búsqueda |
| `radius` | float | `5` | Radio de búsqueda en km |

**Response (`200`):**
```json
{
  "data": [
    {
      "id": "uuid",
      "owner_id": "uuid",
      "name": "Mi Tienda",
      "latitude": 20.6597,
      "longitude": -103.3496
    }
  ]
}
```

---

#### `GET /api/v1/businesses/{id}` 🔒 autenticado

Obtiene un negocio por ID (incluye sus puntos de entrega).

**Response (`200`):**
```json
{
  "data": {
    "id": "uuid",
    "name": "Mi Tienda",
    "delivery_points": [
      {
        "id": "uuid",
        "business_id": "uuid",
        "name": "Entrada principal",
        "latitude": 20.6600,
        "longitude": -103.3500,
        "active": true,
        "created_at": "2026-03-22T20:00:00Z"
      }
    ]
  }
}
```

**Errores:** `404` si no existe.

---

#### `GET /api/v1/businesses/mine` 🔒 seller

Lista los negocios del vendedor autenticado.

**Response (`200`):** Array de negocios (misma estructura).

---

#### `POST /api/v1/businesses/{id}/delivery-points` 🔒 seller

Agrega un punto de entrega a un negocio propio.

**Request:**
```json
{
  "name": "Entrada principal",
  "latitude": 20.6600,
  "longitude": -103.3500
}
```

**Response (`201`):** Objeto `DeliveryPoint`.

**Errores:** `403` si no es dueño del negocio.

---

#### `DELETE /api/v1/businesses/{id}` 🔒 seller

Eliminación total en cascada de un negocio propio.

**Response:** `204 No Content`

**Errores:** `403` si no es dueño.

---

### 📦 Productos

---

#### `POST /api/v1/businesses/{businessId}/products` 🔒 seller

Crea un producto en un negocio.

**Request:**
```json
{
  "name": "Coca-Cola 600ml",
  "description": "Refresco",
  "price": 18.50,
  "stock": 100,
  "image_url": "https://example.com/coca.jpg"
}
```

**Response (`201`):**
```json
{
  "data": {
    "id": "uuid",
    "business_id": "uuid",
    "name": "Coca-Cola 600ml",
    "description": "Refresco",
    "price": 18.50,
    "stock": 100,
    "image_url": "https://example.com/coca.jpg",
    "active": true,
    "created_at": "2026-03-22T20:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z"
  }
}
```

**Errores:** `403` si no es dueño del negocio.

---

#### `GET /api/v1/businesses/{businessId}/products` 🔒 autenticado

Lista todos los productos activos de un negocio.

**Response (`200`):** Array de productos.

---

#### `GET /api/v1/products/{id}` 🔒 autenticado

Obtiene un producto por ID.

**Response (`200`):** Objeto `Product`.

**Errores:** `404` si no existe.

---

#### `PUT /api/v1/products/{id}` 🔒 seller

Actualiza un producto propio.

**Request:**
```json
{
  "name": "Coca-Cola 600ml",
  "description": "Refresco actualizado",
  "price": 20.00,
  "stock": 80,
  "image_url": "https://example.com/coca-new.jpg"
}
```

**Response (`200`):** Producto actualizado.

**Errores:** `403` si no es dueño.

---

#### `DELETE /api/v1/products/{id}` 🔒 seller

Eliminación total de un producto.

**Response:** `204 No Content`

**Errores:** `403` si no es dueño.

---

### 🛒 Órdenes

---

#### `POST /api/v1/orders` 🔒 buyer

Crea un pedido. Verifica stock y lo decrementa atómicamente.

**Request:**
```json
{
  "business_id": "uuid",
  "type": "online",
  "delivery_point_id": "uuid",
  "reservation_hours": 24,
  "items": [
    { "product_id": "uuid", "quantity": 2 },
    { "product_id": "uuid", "quantity": 1 }
  ]
}
```

> - `type`: `"online"` (genera QR inmediato, status=`paid`) o `"reserved"` (status=`reserved`, expira en `reservation_hours` horas, default 24)
> - `delivery_point_id`: Opcional, punto de entrega
> - `reservation_hours`: Solo aplica para `type: "reserved"`

**Response (`201`):**
```json
{
  "data": {
    "id": "uuid",
    "buyer_id": "uuid",
    "business_id": "uuid",
    "type": "online",
    "status": "paid",
    "total": 57.00,
    "qr_code": "base64-encoded-qr",
    "delivery_point_id": "uuid",
    "pickup_deadline": null,
    "created_at": "2026-03-22T20:00:00Z",
    "updated_at": "2026-03-22T20:00:00Z",
    "items": [
      {
        "id": "uuid",
        "order_id": "uuid",
        "product_id": "uuid",
        "quantity": 2,
        "unit_price": 18.50
      }
    ]
  }
}
```

**Errores:**
| Código | Descripción |
|---|---|
| `400` | Body inválido, stock insuficiente, negocio no encontrado |

---

#### `GET /api/v1/orders/{id}` 🔒 autenticado

Obtiene una orden por ID. Solo visible para el comprador o el vendedor del negocio.

**Response (`200`):** Objeto `Order` (incluye `items`).

**Errores:** `404` no encontrada · `403` sin permisos.

---

#### `GET /api/v1/orders/my` 🔒 buyer

Lista todas las órdenes del comprador autenticado.

**Response (`200`):** Array de órdenes.

---

#### `GET /api/v1/businesses/{businessId}/orders` 🔒 seller

Lista todas las órdenes de un negocio propio.

**Response (`200`):** Array de órdenes.

**Errores:** `403` si no es dueño del negocio.

---

#### `POST /api/v1/orders/scan` 🔒 seller

El vendedor escanea un QR para confirmar la entrega. Marca la orden como `delivered`.

**Request:**
```json
{
  "qr_code": "base64-encoded-qr"
}
```

**Response (`200`):** Orden actualizada con `status: "delivered"`.

**Errores:**
| Código | Descripción |
|---|---|
| `404` | QR no corresponde a ninguna orden |
| `403` | No es vendedor de este negocio |
| `410` | Orden expirada (fue cancelada automáticamente) |
| `409` | Transición de status inválida |

---

#### `POST /api/v1/orders/{id}/cancel` 🔒 buyer

Cancela una orden propia (si no ha sido entregada ni cancelada previamente).

**Request:** Sin body.

**Response (`200`):** Orden con `status: "cancelled"`.

**Errores:** `403` no es dueño · `409` status inválido para cancelar.

---

## 🔄 Flujo de Estados de una Orden

```
online:    pending → paid → ready → delivered
                                  ↘ cancelled

reserved:  reserved → ready → delivered
                             ↘ cancelled
```

> Las órdenes `reserved` que pasen su `pickup_deadline` son canceladas automáticamente por un job que corre cada 5 minutos.

## 💰 Comisiones

Cada orden tiene una comisión del **1%** del total (`order.Commission()`).

---

## 🚀 Deploy

```bash
chmod +x deploy.sh
sudo ./deploy.sh
```

El script instala todo desde cero: Go, PostgreSQL + PostGIS, Nginx + SSL, y configura el servicio systemd.

### Comandos Útiles

```bash
# Ver logs de la API
sudo journalctl -u pos-api -f

# Reiniciar la API
sudo systemctl restart pos-api

# Ver estado
sudo systemctl status pos-api
```
