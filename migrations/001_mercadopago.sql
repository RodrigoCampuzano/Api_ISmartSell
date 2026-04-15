-- ============================================================
--  Migración: Mercado Pago Connect
--  Ejecutar sobre la BD pos_app existente
-- ============================================================

-- 1. Nueva columna en payments para guardar el ID de pago de MP
ALTER TABLE payments ADD COLUMN IF NOT EXISTS mp_payment_id TEXT;

-- 2. Ampliar el CHECK de status para soportar 'authorized' y 'cancelled'
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_status_check;
ALTER TABLE payments ADD CONSTRAINT payments_status_check
    CHECK (status IN ('pending','authorized','completed','failed','refunded','cancelled'));

-- 3. Tabla de credenciales OAuth de vendedores con Mercado Pago
CREATE TABLE IF NOT EXISTS seller_mp_credentials (
    user_id       CHAR(36)    NOT NULL PRIMARY KEY,
    access_token  TEXT        NOT NULL,
    refresh_token TEXT        NOT NULL,
    mp_user_id    TEXT        NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_mp_cred_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
