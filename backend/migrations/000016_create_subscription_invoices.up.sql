CREATE TABLE subscription_invoices (
    id                  UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id     UUID                 NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
    outlet_id           UUID                 NOT NULL REFERENCES outlets(id) ON DELETE RESTRICT,
    amount              INTEGER              NOT NULL,
    due_date            DATE                 NOT NULL,
    tripay_reference    VARCHAR(100),
    tripay_payment_url  TEXT,
    status              invoice_status_enum  NOT NULL DEFAULT 'pending',
    paid_at             TIMESTAMP,
    created_at          TIMESTAMP            NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP            NOT NULL DEFAULT now()
);

