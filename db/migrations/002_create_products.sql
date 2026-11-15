-- 002_create_products.sql
-- Module: 03-product (sdd/specs/modules/03-product.md)
-- Applied automatically by the postgres container on first startup
-- (docker-entrypoint-initdb.d runs files in lexical order).

CREATE TABLE IF NOT EXISTS products (
    id                   SERIAL PRIMARY KEY,
    seller_id            INTEGER       NOT NULL REFERENCES users(id),
    photo                TEXT,
    title                VARCHAR(255)  NOT NULL,
    description          TEXT,
    price                NUMERIC(10,2) NOT NULL CHECK (price > 0),
    discount_percentage  INTEGER       NOT NULL DEFAULT 0 CHECK (discount_percentage >= 0 AND discount_percentage <= 80),
    quantity             INTEGER       NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    status               VARCHAR(10)   NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    category             VARCHAR(100),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products(seller_id);
CREATE INDEX IF NOT EXISTS idx_products_status    ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_category  ON products(category);
