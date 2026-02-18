CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id VARCHAR(64) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    dob DATE NOT NULL,
    phone_number VARCHAR(32),
    id_type VARCHAR(16) NOT NULL CHECK (id_type IN ('Passport', 'DL')),
    id_number VARCHAR(128) NOT NULL,
    kyc_level INTEGER NOT NULL CHECK (kyc_level > 0),
    transaction_pin_has VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id VARCHAR(64) NOT NULL REFERENCES users(customer_id) ON DELETE CASCADE,
    account_number VARCHAR(32) NOT NULL UNIQUE,
    currency CHAR(3) NOT NULL CHECK (currency IN ('USD', 'EUR', 'GBP')),
    available_balance NUMERIC(20, 2) NOT NULL DEFAULT 0,
    ledger_balance NUMERIC(20, 2) NOT NULL DEFAULT 0,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN', 'CLOSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_customer_id ON accounts(customer_id);

CREATE TABLE IF NOT EXISTS transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_reference VARCHAR(64) UNIQUE,
    debit_account_number VARCHAR(32) NOT NULL REFERENCES accounts(account_number),
    credit_account_number VARCHAR(32),
    beneficiary_bank_code VARCHAR(16),
    debit_currency CHAR(3) NOT NULL CHECK (debit_currency IN ('USD', 'EUR', 'GBP')),
    credit_currency CHAR(3) NOT NULL CHECK (credit_currency IN ('USD', 'EUR', 'GBP')),
    debit_amount NUMERIC(20, 2) NOT NULL,
    credit_amount NUMERIC(20, 2) NOT NULL,
    ccy_rate NUMERIC(20, 8) NOT NULL,
    charge_amount NUMERIC(20, 2) NOT NULL DEFAULT 0,
    vat_amount NUMERIC(20, 2) NOT NULL DEFAULT 0,
    status VARCHAR(16) NOT NULL CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED')),
    audit_payload TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_payment_reference ON transfers(payment_reference);

CREATE TABLE IF NOT EXISTS rates (
    id BIGSERIAL PRIMARY KEY,
    from_currency CHAR(3) NOT NULL CHECK (from_currency IN ('USD', 'EUR', 'GBP')),
    to_currency CHAR(3) NOT NULL CHECK (to_currency IN ('USD', 'EUR', 'GBP')),
    sell_rate NUMERIC(20, 8) NOT NULL,
    buy_rate NUMERIC(20, 8) NOT NULL,
    rate_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (from_currency, to_currency, rate_date)
);

CREATE TABLE IF NOT EXISTS transient_account_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,
    payment_reference VARCHAR(64) NOT NULL,
    entry_type VARCHAR(16) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    currency CHAR(3) NOT NULL CHECK (currency IN ('USD', 'EUR', 'GBP')),
    amount NUMERIC(20, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tat_transfer_id ON transient_account_transactions(transfer_id);
CREATE INDEX IF NOT EXISTS idx_tat_payment_reference ON transient_account_transactions(payment_reference);
