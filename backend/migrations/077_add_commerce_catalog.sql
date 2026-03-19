CREATE TABLE IF NOT EXISTS commerce_products (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    product_type VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    cover_image TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    sort_order INTEGER NOT NULL DEFAULT 100,
    recommended BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_commerce_products_code_active
    ON commerce_products (code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_commerce_products_status_sort
    ON commerce_products (status, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_product_prices (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES commerce_products(id) ON DELETE CASCADE,
    price_type VARCHAR(32) NOT NULL,
    amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    original_amount NUMERIC(18, 6) NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_commerce_product_prices_product
    ON commerce_product_prices (product_id, enabled, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_product_group_bindings (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES commerce_products(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id),
    grant_type VARCHAR(32) NOT NULL,
    validity_days INTEGER NOT NULL DEFAULT 0,
    priority INTEGER NOT NULL DEFAULT 50,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_commerce_product_group_bindings_unique
    ON commerce_product_group_bindings (product_id, group_id, grant_type)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_model_listings (
    id BIGSERIAL PRIMARY KEY,
    model_key VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    vendor VARCHAR(64) NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 100,
    recommended BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_commerce_model_listings_key_active
    ON commerce_model_listings (model_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_commerce_model_listings_status_sort
    ON commerce_model_listings (status, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_model_product_bindings (
    id BIGSERIAL PRIMARY KEY,
    model_listing_id BIGINT NOT NULL REFERENCES commerce_model_listings(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES commerce_products(id) ON DELETE CASCADE,
    binding_type VARCHAR(32) NOT NULL DEFAULT 'primary',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_commerce_model_product_bindings_unique
    ON commerce_model_product_bindings (model_listing_id, product_id, binding_type)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES commerce_products(id),
    price_id BIGINT NOT NULL REFERENCES commerce_product_prices(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    payment_provider VARCHAR(64) NOT NULL DEFAULT '',
    amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    paid_at TIMESTAMPTZ NULL,
    expired_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_commerce_orders_order_no_active
    ON commerce_orders (order_no)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_commerce_orders_user_status
    ON commerce_orders (user_id, status, payment_status, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_payment_transactions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES commerce_orders(id) ON DELETE CASCADE,
    provider VARCHAR(64) NOT NULL,
    provider_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    request_payload TEXT NOT NULL DEFAULT '',
    callback_payload TEXT NOT NULL DEFAULT '',
    paid_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
    paid_currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_commerce_payment_transactions_order
    ON commerce_payment_transactions (order_id, status, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS commerce_wallet_ledgers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    order_id BIGINT NULL REFERENCES commerce_orders(id) ON DELETE SET NULL,
    direction VARCHAR(16) NOT NULL,
    change_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
    balance_before NUMERIC(18, 6) NOT NULL DEFAULT 0,
    balance_after NUMERIC(18, 6) NOT NULL DEFAULT 0,
    reason_type VARCHAR(64) NOT NULL DEFAULT '',
    reason_detail TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_commerce_wallet_ledgers_user
    ON commerce_wallet_ledgers (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
