-- Add missing columns to email_outbox table for supporting different email types:
-- email_type: Differentiates between 'verification' and 'password_reset' emails
-- reset_link: Stores the password reset link for password reset emails

ALTER TABLE email_outbox
    ADD COLUMN IF NOT EXISTS email_type TEXT DEFAULT 'verification',
    ADD COLUMN IF NOT EXISTS reset_link TEXT;

-- Update existing constraint to include new email_type values
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'email_outbox_email_type_check'
    ) THEN
        ALTER TABLE email_outbox
            DROP CONSTRAINT email_outbox_email_type_check;
    END IF;
    
    ALTER TABLE email_outbox
        ADD CONSTRAINT email_outbox_email_type_check
        CHECK (email_type IN ('verification', 'password_reset'));
EXCEPTION
    WHEN OTHERS THEN
        NULL;
END$$;

-- Create index for email_type lookups
CREATE INDEX IF NOT EXISTS idx_email_outbox_email_type
    ON email_outbox(email_type);
