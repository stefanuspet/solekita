CREATE TABLE attendances (
    id                 UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID                  NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    outlet_id          UUID                  NOT NULL REFERENCES outlets(id) ON DELETE RESTRICT,
    type               attendance_type_enum  NOT NULL,
    selfie_r2_key      TEXT,
    is_selfie_deleted  BOOLEAN               NOT NULL DEFAULT false,
    created_at         TIMESTAMP             NOT NULL DEFAULT now()
);

