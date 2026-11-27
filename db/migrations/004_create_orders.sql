-- 004_create_orders.sql
-- Module: 05-order (sdd/specs/modules/05-order.md)
-- Created in Phase 2 because checkout (04-cart) writes orders.
-- Applied automatically by the postgres container on first startup
-- (docker-entrypoint-initdb.d runs files in lexical order).

CREATE TABLE IF NOT EXISTS orders (
    id          SERIAL PRIMARY KEY,
    cart_id     INTEGER       NOT NULL REFERENCES carts(id),
    buyer_id    INTEGER       NOT NULL REFERENCES users(id),
    total_price NUMERIC(10,2) NOT NULL,
    status      VARCHAR(15)   NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'paid', 'shipped', 'delivered', 'cancelled')),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status   ON orders(status);

CREATE TABLE IF NOT EXISTS order_items (
    id                   SERIAL PRIMARY KEY,
    order_id             INTEGER       NOT NULL REFERENCES orders(id),
    product_id           INTEGER       NOT NULL REFERENCES products(id),
    quantity             INTEGER       NOT NULL CHECK (quantity > 0),
    unit_price_snapshot  NUMERIC(10,2) NOT NULL,
    discount_snapshot    INTEGER       NOT NULL DEFAULT 0
);

-- The product delete guard looks up order items by product.
CREATE INDEX IF NOT EXISTS idx_order_items_order_id   ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
