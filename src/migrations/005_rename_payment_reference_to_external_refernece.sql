DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transfers'
          AND column_name = 'payment_reference'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transfers'
          AND column_name = 'external_refernece'
    ) THEN
        ALTER TABLE transfers RENAME COLUMN payment_reference TO external_refernece;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transient_account_transactions'
          AND column_name = 'payment_reference'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transient_account_transactions'
          AND column_name = 'external_refernece'
    ) THEN
        ALTER TABLE transient_account_transactions RENAME COLUMN payment_reference TO external_refernece;
    END IF;
END $$;

DROP INDEX IF EXISTS idx_transfers_payment_reference;
CREATE INDEX IF NOT EXISTS idx_transfers_external_refernece ON transfers(external_refernece);

DROP INDEX IF EXISTS idx_tat_payment_reference;
CREATE INDEX IF NOT EXISTS idx_tat_external_refernece ON transient_account_transactions(external_refernece);
