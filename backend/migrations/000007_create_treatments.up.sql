CREATE TABLE treatments (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_id  UUID         NOT NULL REFERENCES outlets(id) ON DELETE RESTRICT,
    name       VARCHAR(100) NOT NULL,
    material   VARCHAR(50),
    price      INTEGER      NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMP    NOT NULL DEFAULT now(),
    updated_at TIMESTAMP    NOT NULL DEFAULT now()
);
