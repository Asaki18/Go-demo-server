CREATE TABLE IF NOT EXISTS orders (
    order_uid          TEXT PRIMARY KEY,
    payload            JSONB,
    track_number       TEXT,
    entry              TEXT,
    locale             TEXT,
    internal_signature TEXT,
    customer_id        TEXT,
    delivery_service   TEXT,
    shardkey           TEXT,
    sm_id              INT,
    date_created       TIMESTAMPTZ,
    oof_shard          TEXT,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
