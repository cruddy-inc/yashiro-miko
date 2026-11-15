CREATE SCHEMA IF NOT EXISTS bank;

-- 1. Core customer profile & KYC record
CREATE TABLE bank.customers (
    customer_id UUID DEFAULT gen_random_uuid(),
    CONSTRAINT customers_customer_id_pkey PRIMARY KEY (customer_id),
    customer_type TEXT NOT NULL,
    CONSTRAINT customers_customer_type_check CHECK (customer_type IN ('RETAIL', 'SME')),
    full_name TEXT NOT NULL,
    CONSTRAINT customers_full_name_check CHECK (length(trim(full_name)) BETWEEN 2 AND 120),
    date_of_birth DATE,
    -- ISO 3166-1 alpha-2 uppercase country code (e.g., 'RO', 'DE')
    country CHAR(2) NOT NULL,
    CONSTRAINT customers_country_check CHECK (country ~ '^[A-Z]{2}$'),
    email TEXT,
    CONSTRAINT customers_email_key UNIQUE (email),
    CONSTRAINT customers_email_check CHECK (
        email IS NULL OR (
            length(trim(email)) BETWEEN 6 AND 254
            AND email ~* '^[a-z0-9._+-]{1,64}@([a-z0-9-]{1,63}\.)+([a-z0-9]{2,18})$'
        )
    ),
    -- E.164 international phone number format: '+' followed by 7 to 15 digits
    phone_e164 TEXT,
    CONSTRAINT customers_phone_e164_check CHECK (phone_e164 IS NULL OR phone_e164 ~ '^\+[1-9][0-9]{6,14}$'),
    segment TEXT NOT NULL DEFAULT 'MASS',
    CONSTRAINT customers_segment_check CHECK (segment IN ('MASS', 'AFFLUENT', 'SME')),
    kyc_status TEXT NOT NULL,
    CONSTRAINT customers_kyc_status_check CHECK (kyc_status IN ('PENDING', 'VERIFIED', 'EXPIRED')),
    kyc_risk_rating TEXT NOT NULL DEFAULT 'LOW',
    CONSTRAINT customers_kyc_risk_rating_check CHECK (kyc_risk_rating IN ('LOW', 'MEDIUM', 'HIGH')),
    is_pep BOOLEAN NOT NULL DEFAULT FALSE,
    occupation TEXT,
    CONSTRAINT customers_occupation_check CHECK (occupation IS NULL OR length(trim(occupation)) BETWEEN 2 AND 100),
    expected_monthly_turnover_minor BIGINT,
    CONSTRAINT customers_expected_monthly_turnover_minor_check CHECK (
        expected_monthly_turnover_minor IS NULL OR expected_monthly_turnover_minor >= 0
    ),
    onboarded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    CONSTRAINT customers_status_check CHECK (status IN ('ACTIVE', 'BLOCKED', 'CLOSED'))
);

-- 2. Bank accounts & balances (amounts stored in minor units / cents)
CREATE TABLE bank.accounts (
    account_id UUID DEFAULT gen_random_uuid(),
    CONSTRAINT accounts_account_id_pkey PRIMARY KEY (account_id),
    customer_id UUID NOT NULL,
    CONSTRAINT accounts_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES bank.customers(customer_id) ON DELETE RESTRICT,
    -- Normalized IBAN without spaces: 2-letter country + 2 check digits + 11 to 30 alphanumeric BBAN chars
    iban TEXT NOT NULL,
    CONSTRAINT accounts_iban_key UNIQUE (iban),
    CONSTRAINT accounts_iban_check CHECK (iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'),
    -- ISO 4217 3-letter currency code (e.g., 'EUR', 'RON')
    currency CHAR(3) NOT NULL,
    CONSTRAINT accounts_currency_check CHECK (currency ~ '^[A-Z]{3}$'),
    account_type TEXT NOT NULL,
    CONSTRAINT accounts_account_type_check CHECK (account_type IN ('CURRENT', 'SAVINGS', 'BUSINESS')),
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    CONSTRAINT accounts_status_check CHECK (status IN ('ACTIVE', 'DORMANT', 'FROZEN', 'CLOSED')),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMPTZ,
    CONSTRAINT accounts_closed_at_check CHECK (closed_at IS NULL OR closed_at >= opened_at),
    booked_balance_minor BIGINT NOT NULL DEFAULT 0,
    available_balance_minor BIGINT NOT NULL DEFAULT 0,
    overdraft_limit_minor BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT accounts_overdraft_limit_minor_check CHECK (overdraft_limit_minor >= 0),
    daily_transfer_limit_minor BIGINT NOT NULL DEFAULT 500000,
    CONSTRAINT accounts_daily_transfer_limit_minor_check CHECK (daily_transfer_limit_minor >= 0),
    instant_payment_limit_minor BIGINT NOT NULL DEFAULT 1000000,
    CONSTRAINT accounts_instant_payment_limit_minor_check CHECK (instant_payment_limit_minor >= 0)
);

CREATE INDEX idx_accounts_customer_id ON bank.accounts (customer_id);

-- 3. Customer devices & trust bindings
CREATE TABLE bank.devices (
    customer_id UUID NOT NULL,
    CONSTRAINT devices_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES bank.customers(customer_id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    CONSTRAINT devices_device_id_check CHECK (length(trim(device_id)) BETWEEN 1 AND 128),
    platform TEXT NOT NULL,
    CONSTRAINT devices_platform_check CHECK (platform IN ('IOS', 'ANDROID', 'WEB')),
    os_version TEXT,
    CONSTRAINT devices_os_version_check CHECK (os_version IS NULL OR length(trim(os_version)) <= 32),
    app_version TEXT,
    CONSTRAINT devices_app_version_check CHECK (app_version IS NULL OR length(trim(app_version)) <= 32),
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT devices_last_seen_at_check CHECK (last_seen_at >= first_seen_at),
    bound_at TIMESTAMPTZ,
    is_trusted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT devices_pkey PRIMARY KEY (customer_id, device_id)
);

-- 4. Active & historical login sessions
CREATE TABLE bank.sessions (
    session_id UUID DEFAULT gen_random_uuid(),
    CONSTRAINT sessions_session_id_pkey PRIMARY KEY (session_id),
    customer_id UUID NOT NULL,
    CONSTRAINT sessions_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES bank.customers(customer_id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    CONSTRAINT sessions_device_id_check CHECK (length(trim(device_id)) BETWEEN 1 AND 128),
    channel TEXT NOT NULL,
    CONSTRAINT sessions_channel_check CHECK (channel IN ('MOBILE_APP', 'WEB')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ,
    CONSTRAINT sessions_ended_at_check CHECK (ended_at IS NULL OR ended_at >= started_at),
    ip INET,
    ip_country CHAR(2),
    CONSTRAINT sessions_ip_country_check CHECK (ip_country IS NULL OR ip_country ~ '^[A-Z]{2}$'),
    is_proxy BOOLEAN NOT NULL DEFAULT FALSE,
    auth_method TEXT,
    CONSTRAINT sessions_auth_method_check CHECK (
        auth_method IS NULL OR auth_method IN ('PIN', 'BIOMETRIC', 'PASSWORD_2FA')
    )
);

CREATE INDEX idx_sessions_customer_id_started_at ON bank.sessions (customer_id, started_at DESC);

-- 5. Authentication & security audit events
CREATE TABLE bank.auth_events (
    event_id BIGSERIAL,
    CONSTRAINT auth_events_event_id_pkey PRIMARY KEY (event_id),
    customer_id UUID NOT NULL,
    CONSTRAINT auth_events_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES bank.customers(customer_id) ON DELETE CASCADE,
    device_id TEXT,
    CONSTRAINT auth_events_device_id_check CHECK (device_id IS NULL OR length(trim(device_id)) BETWEEN 1 AND 128),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,
    CONSTRAINT auth_events_event_type_check CHECK (
        event_type IN ('LOGIN_SUCCESS', 'LOGIN_FAILED', 'MFA_FAILED', 'PASSWORD_RESET', 'DEVICE_BOUND')
    ),
    ip INET,
    ip_country CHAR(2),
    CONSTRAINT auth_events_ip_country_check CHECK (ip_country IS NULL OR ip_country ~ '^[A-Z]{2}$')
);

CREATE INDEX idx_auth_events_customer_id_occurred_at ON bank.auth_events (customer_id, occurred_at DESC);

-- 6. Customer profile & limit modification log
CREATE TABLE bank.profile_changes (
    change_id BIGSERIAL,
    CONSTRAINT profile_changes_change_id_pkey PRIMARY KEY (change_id),
    customer_id UUID NOT NULL,
    CONSTRAINT profile_changes_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES bank.customers(customer_id) ON DELETE CASCADE,
    field TEXT NOT NULL,
    CONSTRAINT profile_changes_field_check CHECK (field IN ('PHONE', 'EMAIL', 'ADDRESS', 'LIMIT')),
    old_value_hash TEXT,
    CONSTRAINT profile_changes_old_value_hash_check CHECK (
        old_value_hash IS NULL OR length(trim(old_value_hash)) >= 32
    ),
    new_value_hash TEXT,
    CONSTRAINT profile_changes_new_value_hash_check CHECK (
        new_value_hash IS NULL OR length(trim(new_value_hash)) >= 32
    ),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_id UUID,
    CONSTRAINT profile_changes_session_id_fkey FOREIGN KEY (session_id) REFERENCES bank.sessions(session_id) ON DELETE SET NULL
);

CREATE INDEX idx_profile_changes_customer_id_changed_at ON bank.profile_changes (customer_id, changed_at DESC);

-- 7. Saved payees / beneficiaries & Verification of Payee (VoP) results
CREATE TABLE bank.beneficiaries (
    beneficiary_id UUID DEFAULT gen_random_uuid(),
    CONSTRAINT beneficiaries_beneficiary_id_pkey PRIMARY KEY (beneficiary_id),
    customer_id UUID NOT NULL,
    CONSTRAINT beneficiaries_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES bank.customers(customer_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    CONSTRAINT beneficiaries_name_check CHECK (length(trim(name)) BETWEEN 1 AND 140),
    iban TEXT NOT NULL,
    CONSTRAINT beneficiaries_iban_check CHECK (iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'),
    -- ISO 9362 BIC: 4-char bank + 2-char country + 2-char location + optional 3-char branch (8 or 11 chars)
    bic VARCHAR(11),
    CONSTRAINT beneficiaries_bic_check CHECK (
        bic IS NULL OR bic ~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'
    ),
    nickname TEXT,
    CONSTRAINT beneficiaries_nickname_check CHECK (nickname IS NULL OR length(trim(nickname)) BETWEEN 1 AND 50),
    added_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    added_session_id UUID,
    CONSTRAINT beneficiaries_added_session_id_fkey FOREIGN KEY (added_session_id) REFERENCES bank.sessions(session_id) ON DELETE SET NULL,
    vop_result TEXT,
    CONSTRAINT beneficiaries_vop_result_check CHECK (
        vop_result IS NULL OR vop_result IN ('MATCH', 'CLOSE_MATCH', 'NO_MATCH', 'NOT_POSSIBLE')
    ),
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    CONSTRAINT beneficiaries_status_check CHECK (status IN ('ACTIVE', 'REMOVED')),
    last_used_at TIMESTAMPTZ,
    CONSTRAINT beneficiaries_customer_id_iban_key UNIQUE (customer_id, iban)
);

-- 8. Outgoing payment initiation orders
CREATE TABLE bank.payment_orders (
    payment_id UUID DEFAULT gen_random_uuid(),
    CONSTRAINT payment_orders_payment_id_pkey PRIMARY KEY (payment_id),
    debtor_account_id UUID NOT NULL,
    CONSTRAINT payment_orders_debtor_account_id_fkey FOREIGN KEY (debtor_account_id) REFERENCES bank.accounts(account_id) ON DELETE RESTRICT,
    creditor_name TEXT NOT NULL,
    CONSTRAINT payment_orders_creditor_name_check CHECK (length(trim(creditor_name)) BETWEEN 1 AND 140),
    creditor_iban TEXT NOT NULL,
    CONSTRAINT payment_orders_creditor_iban_check CHECK (creditor_iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'),
    creditor_bic VARCHAR(11),
    CONSTRAINT payment_orders_creditor_bic_check CHECK (
        creditor_bic IS NULL OR creditor_bic ~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'
    ),
    amount_minor BIGINT NOT NULL,
    CONSTRAINT payment_orders_amount_minor_check CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    CONSTRAINT payment_orders_currency_check CHECK (currency ~ '^[A-Z]{3}$'),
    scheme TEXT NOT NULL,
    CONSTRAINT payment_orders_scheme_check CHECK (scheme IN ('SEPA_INSTANT', 'SEPA_CT', 'INTERNAL')),
    -- SEPA EndToEndId max length is 35 characters
    end_to_end_id TEXT NOT NULL,
    CONSTRAINT payment_orders_end_to_end_id_check CHECK (length(trim(end_to_end_id)) BETWEEN 1 AND 35),
    -- SEPA unstructured remittance info max length is 140 characters
    remittance_info TEXT,
    CONSTRAINT payment_orders_remittance_info_check CHECK (
        remittance_info IS NULL OR length(remittance_info) <= 140
    ),
    session_id UUID,
    CONSTRAINT payment_orders_session_id_fkey FOREIGN KEY (session_id) REFERENCES bank.sessions(session_id) ON DELETE SET NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,
    CONSTRAINT payment_orders_status_check CHECK (
        status IN ('RECEIVED', 'SCREENING', 'HELD', 'EXECUTED', 'REJECTED', 'RETURNED')
    ),
    status_reason TEXT,
    CONSTRAINT payment_orders_status_reason_check CHECK (
        status_reason IS NULL OR length(trim(status_reason)) <= 255
    ),
    executed_at TIMESTAMPTZ,
    CONSTRAINT payment_orders_executed_at_check CHECK (executed_at IS NULL OR executed_at >= requested_at)
);

CREATE INDEX idx_payment_orders_debtor_account_id_requested_at ON bank.payment_orders (debtor_account_id, requested_at DESC);

-- 9. Settled account statement ledger entries
CREATE TABLE bank.transactions (
    txn_id UUID DEFAULT gen_random_uuid(),
    CONSTRAINT transactions_txn_id_pkey PRIMARY KEY (txn_id),
    account_id UUID NOT NULL,
    CONSTRAINT transactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES bank.accounts(account_id) ON DELETE RESTRICT,
    payment_id UUID,
    CONSTRAINT transactions_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES bank.payment_orders(payment_id) ON DELETE SET NULL,
    booked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    booking_date DATE NOT NULL,
    value_date DATE NOT NULL,
    direction TEXT NOT NULL,
    CONSTRAINT transactions_direction_check CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount_minor BIGINT NOT NULL,
    CONSTRAINT transactions_amount_minor_check CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    CONSTRAINT transactions_currency_check CHECK (currency ~ '^[A-Z]{3}$'),
    balance_after_minor BIGINT NOT NULL,
    scheme TEXT NOT NULL,
    CONSTRAINT transactions_scheme_check CHECK (
        scheme IN ('SEPA_INSTANT', 'SEPA_CT', 'SEPA_DD', 'CARD', 'ATM', 'FEE', 'INTEREST', 'INTERNAL')
    ),
    counterparty_name TEXT,
    CONSTRAINT transactions_counterparty_name_check CHECK (
        counterparty_name IS NULL OR length(trim(counterparty_name)) BETWEEN 1 AND 140
    ),
    counterparty_iban TEXT,
    CONSTRAINT transactions_counterparty_iban_check CHECK (
        counterparty_iban IS NULL OR counterparty_iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'
    ),
    counterparty_bic VARCHAR(11),
    CONSTRAINT transactions_counterparty_bic_check CHECK (
        counterparty_bic IS NULL OR counterparty_bic ~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'
    ),
    remittance_info TEXT,
    CONSTRAINT transactions_remittance_info_check CHECK (
        remittance_info IS NULL OR length(remittance_info) <= 140
    ),
    end_to_end_id TEXT,
    CONSTRAINT transactions_end_to_end_id_check CHECK (
        end_to_end_id IS NULL OR length(trim(end_to_end_id)) BETWEEN 1 AND 35
    ),
    channel TEXT,
    CONSTRAINT transactions_channel_check CHECK (
        channel IS NULL OR channel IN ('MOBILE_APP', 'WEB', 'API', 'BRANCH', 'ATM', 'POS')
    ),
    status TEXT NOT NULL DEFAULT 'BOOKED',
    CONSTRAINT transactions_status_check CHECK (status IN ('PENDING', 'BOOKED', 'REVERSED'))
);

CREATE INDEX idx_transactions_account_id_booked_at ON bank.transactions (account_id, booked_at DESC);
CREATE INDEX idx_transactions_counterparty_iban_booked_at ON bank.transactions (counterparty_iban, booked_at DESC);