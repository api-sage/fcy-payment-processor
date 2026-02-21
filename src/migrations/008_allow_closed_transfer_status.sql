ALTER TABLE transfers
    DROP CONSTRAINT IF EXISTS transfers_status_check;

ALTER TABLE transfers
    ADD CONSTRAINT transfers_status_check
    CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED', 'CLOSED'));
