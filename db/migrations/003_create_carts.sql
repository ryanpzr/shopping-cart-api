-- 003_create_carts.sql
-- Module: 04-cart (sdd/specs/modules/04-cart.md)
-- Applied automatically by the postgres container on first startup
-- (docker-entrypoint-initdb.d runs files in lexical order).

CREATE TABLE IF NOT EXISTS carts (
    id         SERIAL PRIMARY KEY,
    buyer_id   INTEGER     NOT NULL REFERENCES users(id),
    status     VARCHAR(15) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'checked_out')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_carts_buyer_id ON carts(buyer_id);

-- At most one active cart per buyer, even under concurrent requests.
CREATE UNIQUE INDEX IF NOT EXISTS ux_carts_one_active_per_buyer ON carts(buyer_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS cart_items (
    id             SERIAL PRIMARY KEY,
    cart_id        INTEGER       NOT NULL REFERENCES carts(id),
    -- Deleting a product removes it from carts; purchase history lives in order_items.
    product_id     INTEGER       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity       INTEGER       NOT NULL CHECK (quantity > 0),
    price_snapshot NUMERIC(10,2) NOT NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);
