CREATE TABLE subscriptions (
    id                      UUID                   PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id               UUID                   NOT NULL UNIQUE REFERENCES outlets(id) ON DELETE RESTRICT,
    plan                    subscription_plan_enum NOT NULL DEFAULT 'monthly',
    price_per_month         INTEGER                NOT NULL DEFAULT 29000,
    trial_started_at        TIMESTAMP,
    trial_ends_at           TIMESTAMP,
    subscription_started_at DATE,
    next_due_date           DATE,
    suspended_at            TIMESTAMP,
    inactive_at             TIMESTAMP,
    created_at              TIMESTAMP              NOT NULL DEFAULT now(),
    updated_at              TIMESTAMP              NOT NULL DEFAULT now()
);
