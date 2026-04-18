-- Ubah fk_outlets_owner menjadi DEFERRABLE INITIALLY DEFERRED
-- untuk menyelesaikan circular dependency saat register:
-- outlets.owner_id → users, users.outlet_id → outlets
-- Dengan DEFERRABLE, constraint dicek saat COMMIT bukan saat INSERT.

ALTER TABLE outlets DROP CONSTRAINT fk_outlets_owner;

ALTER TABLE outlets
    ADD CONSTRAINT fk_outlets_owner
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;
