CREATE SCHEMA IF NOT EXISTS yaotose;

CREATE TABLE yaotose.blacklist (
    tenant_id TEXT NOT NULL,
    CONSTRAINT blacklist_tenant_id_check CHECK (length(trim(tenant_id)) BETWEEN 1 AND 64),
    iban_hash TEXT NOT NULL,
    CONSTRAINT blacklist_iban_hash_check CHECK (iban_hash ~ '^[0-9a-f]{64}$'),
    key_id TEXT NOT NULL DEFAULT 'k1',
    CONSTRAINT blacklist_key_id_check CHECK (length(trim(key_id)) BETWEEN 1 AND 32),
    reason TEXT[] NOT NULL,
    ai_message TEXT NOT NULL,
    CONSTRAINT blacklist_ai_message_check CHECK (length(trim(ai_message)) > 0),
    blacklisted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    CONSTRAINT blacklist_expires_at_check CHECK (expires_at IS NULL OR expires_at > blacklisted_at),
    CONSTRAINT blacklist_pkey PRIMARY KEY (tenant_id, iban_hash)
);

CREATE INDEX idx_blacklist_expires_at ON yaotose.blacklist(expires_at DESC);

CREATE TABLE yaotose.payer_payee (
    tenant_id TEXT NOT NULL,
    CONSTRAINT payer_payee_tenant_id_check CHECK (length(trim(tenant_id)) BETWEEN 1 AND 64),
    payer_hash TEXT NOT NULL,
    CONSTRAINT payer_payee_payer_hash_check CHECK (payer_hash ~ '^[0-9a-f]{64}'),
    payee_hash TEXT NOT NULL,
    CONSTRAINT payer_payee_payee_hash_check CHECK (payee_hash ~ '^[0-9a-f]{64}'),
    key_id TEXT NOT NULL DEFAULT 'k1',
    n_payments INT NOT NULL DEFAULT 1,
    CONSTRAINT payer_payee_n_payments_check CHECK (n_payments > 0),
    total_amount BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT payer_payee_total_amount_check CHECK (total_amount > 0),
    min_amount BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT payer_payee_min_amount_check CHECK (min_amount > 0),
    max_amount BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT payer_payee_max_amount_check CHECK (max_amount > 0),
    first_transaction TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_transaction TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT payer_payee_pkey PRIMARY KEY (tenant_id, payer_hash, payee_hash)
);