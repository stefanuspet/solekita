CREATE TABLE outlets (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name                   VARCHAR(100) NOT NULL,
    code                   VARCHAR(5)   NOT NULL UNIQUE,
    address                TEXT,
    phone                  VARCHAR(20),
    owner_id               UUID         NOT NULL,
    subscription_status    subscription_status_enum NOT NULL DEFAULT 'trial',
    overdue_threshold_days INTEGER      NOT NULL DEFAULT 7,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    created_at             TIMESTAMP    NOT NULL DEFAULT now(),
    updated_at             TIMESTAMP    NOT NULL DEFAULT now()
);

-- FK owner_id → users ditambahkan di migration 000003_create_users
-- karena users.outlet_id → outlets (circular dependency)
