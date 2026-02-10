CREATE TABLE IF NOT EXISTS email_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    to_email TEXT NOT NULL,
    to_name TEXT NOT NULL DEFAULT '',
    verify_link TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'email_outbox_status_check'
    ) THEN
        ALTER TABLE email_outbox
            ADD CONSTRAINT email_outbox_status_check
            CHECK (status IN ('pending', 'processing', 'retry', 'sent', 'failed'));
    END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_email_outbox_status_next_attempt
    ON email_outbox(status, next_attempt_at);

CREATE INDEX IF NOT EXISTS idx_email_outbox_created_at
    ON email_outbox(created_at DESC);
