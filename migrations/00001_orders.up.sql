CREATE SCHEMA IF NOT EXISTS orders;

create table if not exists orders.orders (
    order_id UUID PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL UNIQUE,
    entry VARCHAR(125) NOT NULL,
    locale VARCHAR(125) NOT NULL,
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255) NOT NULL,
    delivery_service VARCHAR(255),
    shardkey VARCHAR(125) NOT NULL,
    sm_id BIGINT,
    date_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    oof_shard VARCHAR(125) NOT NULL
);

create index if not exists idx_orders_customer_id on orders.orders(customer_id);

CREATE TABLE IF NOT EXISTS orders.delivery (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE REFERENCES orders.orders(order_id),
    name TEXT NOT NULL,
    phone VARCHAR(25) NOT NULL,
    zip VARCHAR(125) NOT NULL,
    city VARCHAR(125) NOT NULL,
    address VARCHAR(125) NOT NULL,
    region VARCHAR(125),
    email VARCHAR(125) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_delivery_order_id ON orders.delivery(order_id);

create table if not exists orders.items (
    id BIGSERIAL PRIMARY KEY,
    chrt_id BIGINT, -- NEED TO BE UNIQUE?
    track_number VARCHAR(255) NOT NULL,
    price NUMERIC(15, 2) NOT NULL,
    rid VARCHAR(255) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    sale BIGINT,
    size VARCHAR(125),
    total_price NUMERIC(15, 2) NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR(125) NOT NULL,
    status INT NOT NULL
);

create table if not exists orders.order_items (
    order_id UUID NOT NULL REFERENCES orders.orders(order_id),
    item_id BIGINT NOT NULL REFERENCES orders.items(id),
    PRIMARY KEY (order_id, item_id)
);