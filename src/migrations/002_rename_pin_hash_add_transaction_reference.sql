DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
          AND column_name = 'transaction_pin_has'
    ) THEN
        ALTER TABLE users RENAME COLUMN transaction_pin_has TO transaction_pin_hash;
    END IF;
END $$;

ALTER TABLE transfers
    ADD COLUMN IF NOT EXISTS transaction_reference VARCHAR(64);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transfers_transaction_reference
    ON transfers(transaction_reference)
    WHERE transaction_reference IS NOT NULL;
