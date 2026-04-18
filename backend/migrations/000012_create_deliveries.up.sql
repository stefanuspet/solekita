CREATE TABLE deliveries (
    id               UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         UUID                  NOT NULL UNIQUE REFERENCES orders(id) ON DELETE RESTRICT,
    courier_id       UUID                  REFERENCES users(id) ON DELETE SET NULL,
    pickup_address   TEXT,
    delivery_address TEXT,
    pickup_status    pickup_status_enum    NOT NULL DEFAULT 'pending',
    delivery_status  delivery_status_enum  NOT NULL DEFAULT 'pending',
    pickup_notes     TEXT,
    delivery_notes   TEXT,
    picked_up_at     TIMESTAMP,
    delivered_at     TIMESTAMP,
    created_at       TIMESTAMP             NOT NULL DEFAULT now(),
    updated_at       TIMESTAMP             NOT NULL DEFAULT now()
);

