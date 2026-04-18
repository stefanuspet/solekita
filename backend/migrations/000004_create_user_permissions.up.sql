CREATE TABLE user_permissions (
    id         UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission permission_enum NOT NULL,
    created_at TIMESTAMP       NOT NULL DEFAULT now()
);

