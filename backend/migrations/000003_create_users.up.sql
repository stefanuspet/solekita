CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id     UUID         NOT NULL REFERENCES outlets(id) ON DELETE RESTRICT,
    name          VARCHAR(100) NOT NULL,
    phone         VARCHAR(20)  NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    is_owner      BOOLEAN      NOT NULL DEFAULT false,
    is_active     BOOLEAN      NOT NULL DEFAULT true,
    last_login_at TIMESTAMP,
    created_at    TIMESTAMP    NOT NULL DEFAULT now(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT now()
);

-- Selesaikan circular dependency: tambahkan FK owner_id di outlets → users
ALTER TABLE outlets
    ADD CONSTRAINT fk_outlets_owner
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;
