-- ============================================================
--  POS App — Schema PostgreSQL + PostGIS
-- ============================================================

CREATE EXTENSION IF NOT EXISTS postgis;

-- ------------------------------------------------------------
-- users
-- ------------------------------------------------------------
CREATE TABLE users (
    id         CHAR(36)                   NOT NULL PRIMARY KEY,
    name       VARCHAR(100)               NOT NULL,
    email      VARCHAR(150)               NOT NULL UNIQUE,
    password   VARCHAR(255)               NOT NULL,
    role       VARCHAR(10)                NOT NULL CHECK (role IN ('seller','buyer')),
    active     BOOLEAN                    NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ                NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------
-- businesses
-- ------------------------------------------------------------
CREATE TABLE businesses (
    id          CHAR(36)                   NOT NULL PRIMARY KEY,
    owner_id    CHAR(36)                   NOT NULL,
    name        VARCHAR(150)               NOT NULL,
    description TEXT,
    type        VARCHAR(100),
    location    GEOGRAPHY(Point, 4326)     NOT NULL,
    active      BOOLEAN                    NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_business_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_businesses_owner   ON businesses(owner_id);
CREATE INDEX idx_businesses_location ON businesses USING GIST(location);

-- ------------------------------------------------------------
-- delivery_points
-- ------------------------------------------------------------
CREATE TABLE delivery_points (
    id          CHAR(36)                   NOT NULL PRIMARY KEY,
    business_id CHAR(36)                   NOT NULL,
    name        VARCHAR(150)               NOT NULL,
    location    GEOGRAPHY(Point, 4326)     NOT NULL,
    active      BOOLEAN                    NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_dp_business FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE
);

CREATE INDEX idx_dp_business ON delivery_points(business_id);

-- ------------------------------------------------------------
-- products
-- ------------------------------------------------------------
CREATE TABLE products (
    id          CHAR(36)      NOT NULL PRIMARY KEY,
    business_id CHAR(36)      NOT NULL,
    name        VARCHAR(150)  NOT NULL,
    description TEXT,
    price       DECIMAL(10,2) NOT NULL,
    stock       INT           NOT NULL DEFAULT 0,
    image_url   VARCHAR(500),
    active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_product_business FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE
);

CREATE INDEX idx_products_business ON products(business_id);

-- ------------------------------------------------------------
-- orders
-- ------------------------------------------------------------
CREATE TABLE orders (
    id                CHAR(36)      NOT NULL PRIMARY KEY,
    buyer_id          CHAR(36)      NOT NULL,
    business_id       CHAR(36)      NOT NULL,
    type              VARCHAR(10)   NOT NULL CHECK (type IN ('online','reserved')),
    status            VARCHAR(12)   NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending','paid','reserved','ready','delivered','cancelled')),
    total             DECIMAL(10,2) NOT NULL,
    qr_code           VARCHAR(500),
    delivery_point_id CHAR(36),
    pickup_deadline   TIMESTAMPTZ,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_order_buyer    FOREIGN KEY (buyer_id)          REFERENCES users(id),
    CONSTRAINT fk_order_business FOREIGN KEY (business_id)       REFERENCES businesses(id),
    CONSTRAINT fk_order_dp       FOREIGN KEY (delivery_point_id) REFERENCES delivery_points(id)
);

CREATE INDEX idx_orders_buyer    ON orders(buyer_id);
CREATE INDEX idx_orders_business ON orders(business_id);
CREATE INDEX idx_orders_status   ON orders(status);

-- ------------------------------------------------------------
-- order_items
-- ------------------------------------------------------------
CREATE TABLE order_items (
    id         CHAR(36)      NOT NULL PRIMARY KEY,
    order_id   CHAR(36)      NOT NULL,
    product_id CHAR(36)      NOT NULL,
    quantity   INT           NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    CONSTRAINT fk_item_order   FOREIGN KEY (order_id)   REFERENCES orders(id)   ON DELETE CASCADE,
    CONSTRAINT fk_item_product FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

-- ------------------------------------------------------------
-- payments  (comisión 1% por venta)
-- ------------------------------------------------------------
CREATE TABLE payments (
    id         CHAR(36)      NOT NULL PRIMARY KEY,
    order_id   CHAR(36)      NOT NULL UNIQUE,
    amount     DECIMAL(10,2) NOT NULL,
    commission DECIMAL(10,2) NOT NULL, -- 1% del monto
    method     VARCHAR(10)   NOT NULL DEFAULT 'online' CHECK (method IN ('online','cash')),
    status     VARCHAR(10)   NOT NULL DEFAULT 'pending'
               CHECK (status IN ('pending','completed','failed','refunded')),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_payment_order FOREIGN KEY (order_id) REFERENCES orders(id)
);
