-- Create notification schedules table for push reminders
CREATE TABLE IF NOT EXISTS notification_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Schedule configuration
    enabled BOOLEAN DEFAULT true,
    timezone TEXT DEFAULT 'UTC',
    
    -- Days of week: array of integers 0-6 (Mon-Sun) or NULL for everyday
    days_of_week INT[] DEFAULT'{0,1,2,3,4,5,6}',
    
    -- Time of day for reminder in HH:MM format
    reminder_time TEXT NOT NULL DEFAULT '09:00',
    
    -- Notification type
    notification_type TEXT DEFAULT 'start_session', -- start_session | daily_summary | motivational
    
    -- Last sent timestamp to avoid duplicate sends
    last_sent_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_schedules_user_id 
    ON notification_schedules(user_id);

CREATE INDEX IF NOT EXISTS idx_notification_schedules_enabled 
    ON notification_schedules(enabled);
