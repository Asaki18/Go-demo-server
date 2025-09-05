CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    payload   JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
