CREATE TABLE photos (
    id           UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID            NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    type         photo_type_enum NOT NULL,
    r2_key       TEXT            NOT NULL,
    file_size_kb INTEGER,
    is_deleted   BOOLEAN         NOT NULL DEFAULT false,
    uploaded_at  TIMESTAMP       NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMP
);

