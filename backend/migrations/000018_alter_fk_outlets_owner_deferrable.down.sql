ALTER TABLE outlets DROP CONSTRAINT fk_outlets_owner;

ALTER TABLE outlets
    ADD CONSTRAINT fk_outlets_owner
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT;
