CREATE TABLE order_sequences (
    outlet_id UUID NOT NULL REFERENCES outlets(id) ON DELETE RESTRICT,
    date      DATE NOT NULL,
    last_seq  INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (outlet_id, date)
);
