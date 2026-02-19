DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'transfers'
          AND column_name = 'ccy_rate'
    ) THEN
        ALTER TABLE public.transfers RENAME COLUMN ccy_rate TO fcy_rate;
    END IF;
END $$;
