CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(255) NOT NULL,
    sku VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    CHECK (quantity >= 0)
);