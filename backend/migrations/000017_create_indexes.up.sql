-- ── Full Indexes ──────────────────────────────────────────────────────────────

-- orders
CREATE INDEX idx_orders_outlet_created ON orders (outlet_id, created_at DESC);
CREATE INDEX idx_orders_outlet_status  ON orders (outlet_id, status);
CREATE INDEX idx_orders_customer       ON orders (customer_id);
CREATE INDEX idx_orders_kasir          ON orders (kasir_id);
CREATE INDEX idx_orders_overdue        ON orders (outlet_id, status, updated_at);

-- photos
CREATE INDEX idx_photos_order ON photos (order_id);

-- payments
CREATE INDEX idx_payments_order ON payments (order_id);

-- attendances
CREATE INDEX idx_attendances_user   ON attendances (user_id, created_at DESC);
CREATE INDEX idx_attendances_outlet ON attendances (outlet_id, created_at DESC);

-- customers
CREATE UNIQUE INDEX idx_customers_outlet_phone ON customers (outlet_id, phone);

-- user_permissions
CREATE UNIQUE INDEX idx_user_permissions_unique ON user_permissions (user_id, permission);

-- deliveries
CREATE INDEX idx_deliveries_courier ON deliveries (courier_id, created_at DESC);

-- order_logs
CREATE INDEX idx_order_logs_order ON order_logs (order_id, created_at DESC);

-- subscription_invoices
CREATE INDEX idx_invoices_outlet_due ON subscription_invoices (outlet_id, due_date DESC);
CREATE INDEX idx_invoices_status_due ON subscription_invoices (status, due_date);

-- refresh_tokens
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens (token_hash);

-- ── Partial Indexes ───────────────────────────────────────────────────────────

-- Order aktif (belum selesai/dibatalkan) — dipakai kasir
CREATE INDEX idx_orders_active ON orders (outlet_id, created_at DESC)
WHERE status NOT IN ('diambil', 'diantar', 'dibatalkan');

-- Order menggantung (selesai tapi belum diambil/diantar) — dipakai cron job & dashboard
CREATE INDEX idx_orders_pending_pickup ON orders (outlet_id, updated_at)
WHERE status = 'selesai';

-- Invoice belum dibayar — dipakai cron job billing harian
CREATE INDEX idx_invoices_unpaid ON subscription_invoices (due_date)
WHERE status = 'pending';

-- Outlet masih trial — dipakai cron job trial reminder
CREATE INDEX idx_subscriptions_trial_active ON subscriptions (trial_ends_at)
WHERE trial_ends_at IS NOT NULL AND subscription_started_at IS NULL;

-- Outlet berbayar yang akan jatuh tempo — dipakai cron job suspend
CREATE INDEX idx_subscriptions_due ON subscriptions (next_due_date)
WHERE next_due_date IS NOT NULL;

-- ── Trigger: auto-update updated_at ──────────────────────────────────────────

CREATE OR REPLACE FUNCTION fn_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_outlets_updated_at
    BEFORE UPDATE ON outlets
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_treatments_updated_at
    BEFORE UPDATE ON treatments
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_deliveries_updated_at
    BEFORE UPDATE ON deliveries
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

CREATE TRIGGER trg_subscription_invoices_updated_at
    BEFORE UPDATE ON subscription_invoices
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

-- ── Trigger: auto-update total_orders di customers ───────────────────────────

CREATE OR REPLACE FUNCTION fn_update_customer_total_orders()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.status != 'dibatalkan' THEN
        UPDATE customers
        SET total_orders = total_orders + 1,
            last_order_at = now()
        WHERE id = NEW.customer_id;

    ELSIF TG_OP = 'UPDATE'
        AND OLD.status != 'dibatalkan'
        AND NEW.status = 'dibatalkan' THEN
        UPDATE customers
        SET total_orders = GREATEST(total_orders - 1, 0)
        WHERE id = NEW.customer_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_orders_customer_counter
    AFTER INSERT OR UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION fn_update_customer_total_orders();
