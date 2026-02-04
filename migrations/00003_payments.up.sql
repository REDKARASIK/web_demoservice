CREATE TABLE IF NOT EXISTS orders.payments (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE REFERENCES orders.orders(order_id),
    transaction TEXT UNIQUE NOT NULL,
    request_id VARCHAR(125) UNIQUE,
    currency VARCHAR(25) NOT NULL,
    provider VARCHAR(125) NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    payment_dt BIGINT NOT NULL,
    bank BIGINT REFERENCES banks.banks(id),
    delivery_cost NUMERIC(15, 2) NOT NULL,
    goods_total BIGINT NOT NULL,
    custom_fee NUMERIC(15, 2) NOT NULL
);