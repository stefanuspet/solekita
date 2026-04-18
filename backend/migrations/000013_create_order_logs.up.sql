CREATE TABLE order_logs (
    id         UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID                 NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    user_id    UUID                 NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action     order_action_enum    NOT NULL,
    old_value  TEXT,
    new_value  TEXT,
    notes      TEXT,
    created_at TIMESTAMP            NOT NULL DEFAULT now()
);

