ALTER TABLE rates
    ADD COLUMN IF NOT EXISTS rate NUMERIC(20, 8);

UPDATE rates
SET rate = COALESCE(rate, sell_rate, buy_rate)
WHERE rate IS NULL;

ALTER TABLE rates
    ALTER COLUMN rate SET NOT NULL;

ALTER TABLE rates
    DROP COLUMN IF EXISTS sell_rate;

ALTER TABLE rates
    DROP COLUMN IF EXISTS buy_rate;
