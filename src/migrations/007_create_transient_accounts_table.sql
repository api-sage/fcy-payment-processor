CREATE TABLE IF NOT EXISTS transient_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(32) NOT NULL UNIQUE,
    account_description VARCHAR(255) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'FCY',
    available_balance NUMERIC(20, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transient_accounts_account_number
    ON transient_accounts(account_number);
