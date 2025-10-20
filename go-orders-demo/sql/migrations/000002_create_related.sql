CREATE TABLE IF NOT EXISTS deliveries (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT, phone TEXT, zip TEXT, city TEXT, address TEXT, region TEXT, email TEXT
);

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT UNIQUE,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount INT,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    chrt_id INT UNIQUE,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    track_number TEXT,
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id INT,
    brand TEXT,
    status INT
);
