CREATE TABLE customers (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id        UUID         NOT NULL REFERENCES outlets(id) ON DELETE RESTRICT,
    name             VARCHAR(100) NOT NULL,
    phone            VARCHAR(20)  NOT NULL,
    total_orders     INTEGER      NOT NULL DEFAULT 0,
    last_order_at    TIMESTAMP,
    notes            TEXT,
    is_blacklisted   BOOLEAN      NOT NULL DEFAULT false,
    blacklist_reason TEXT,
    created_at       TIMESTAMP    NOT NULL DEFAULT now(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT now()
);

