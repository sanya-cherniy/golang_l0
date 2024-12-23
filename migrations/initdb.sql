
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    transaction VARCHAR(100),
    request_id VARCHAR(100),
    currency VARCHAR(10),
    provider VARCHAR(100),
    amount INT,
    payment_dt INT,
    bank VARCHAR(100),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    phone VARCHAR(50),
    zip VARCHAR(20),
    city VARCHAR(100),
    address VARCHAR(255),
    region VARCHAR(100),
    email VARCHAR(100)
);

CREATE TABLE orders (
    order_uid VARCHAR(100) PRIMARY KEY,
    track_number VARCHAR(100),
    entry VARCHAR(100),
    delivery_id INT,
    payment_id INT,
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(100),
    delivery_service VARCHAR(100),
    shardkey VARCHAR(50),
    sm_id INT,
    date_created VARCHAR(255),
    oof_shard VARCHAR(50),
    FOREIGN KEY (delivery_id) REFERENCES deliveries(id),
    FOREIGN KEY (payment_id) REFERENCES payments(id)
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(100),
    chrt_id INT,
    track_number VARCHAR(100),
    price INT,
    rid VARCHAR(100),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(50),
    total_price INT,
    nm_id INT,
    brand VARCHAR(100),
    status INT,
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);
