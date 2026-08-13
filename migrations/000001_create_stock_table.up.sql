CREATE TABLE IF NOT EXISTS stock (
    sku VARCHAR(255) PRIMARY KEY,
    quantity BIGINT NOT NULL CHECK (quantity >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO stock (sku, quantity) VALUES ('iphone-15', 1);
INSERT INTO stock (sku, quantity) VALUES ('case-silicone', 2);