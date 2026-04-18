CREATE TABLE payments (
    id         UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID                  NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    amount     INTEGER               NOT NULL,
    method     payment_method_enum   NOT NULL,
    status     payment_status_enum   NOT NULL DEFAULT 'paid',
    notes      TEXT,
    paid_at    TIMESTAMP,
    created_at TIMESTAMP             NOT NULL DEFAULT now()
);

