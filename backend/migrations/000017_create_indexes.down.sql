-- ── Triggers ─────────────────────────────────────────────────────────────────

DROP TRIGGER IF EXISTS trg_orders_customer_counter ON orders;
DROP FUNCTION IF EXISTS fn_update_customer_total_orders;

DROP TRIGGER IF EXISTS trg_subscription_invoices_updated_at ON subscription_invoices;
DROP TRIGGER IF EXISTS trg_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS trg_deliveries_updated_at ON deliveries;
DROP TRIGGER IF EXISTS trg_treatments_updated_at ON treatments;
DROP TRIGGER IF EXISTS trg_customers_updated_at ON customers;
DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TRIGGER IF EXISTS trg_outlets_updated_at ON outlets;
DROP FUNCTION IF EXISTS fn_update_updated_at;

-- ── Partial Indexes ───────────────────────────────────────────────────────────

DROP INDEX IF EXISTS idx_subscriptions_due;
DROP INDEX IF EXISTS idx_subscriptions_trial_active;
DROP INDEX IF EXISTS idx_invoices_unpaid;
DROP INDEX IF EXISTS idx_orders_pending_pickup;
DROP INDEX IF EXISTS idx_orders_active;

-- ── Full Indexes ──────────────────────────────────────────────────────────────

DROP INDEX IF EXISTS idx_refresh_tokens_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user;
DROP INDEX IF EXISTS idx_invoices_status_due;
DROP INDEX IF EXISTS idx_invoices_outlet_due;
DROP INDEX IF EXISTS idx_order_logs_order;
DROP INDEX IF EXISTS idx_deliveries_courier;
DROP INDEX IF EXISTS idx_user_permissions_unique;
DROP INDEX IF EXISTS idx_customers_outlet_phone;
DROP INDEX IF EXISTS idx_attendances_outlet;
DROP INDEX IF EXISTS idx_attendances_user;
DROP INDEX IF EXISTS idx_payments_order;
DROP INDEX IF EXISTS idx_photos_order;
DROP INDEX IF EXISTS idx_orders_overdue;
DROP INDEX IF EXISTS idx_orders_kasir;
DROP INDEX IF EXISTS idx_orders_customer;
DROP INDEX IF EXISTS idx_orders_outlet_status;
DROP INDEX IF EXISTS idx_orders_outlet_created;
