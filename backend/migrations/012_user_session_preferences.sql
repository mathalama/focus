-- Create user session preferences table
CREATE TABLE IF NOT EXISTS user_session_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Session duration presets in minutes
    preset_durations INT[] DEFAULT '{25,45,90}',
    default_duration INT DEFAULT 25,
    
    -- Break settings in minutes
    short_break_duration INT DEFAULT 5,
    long_break_duration INT DEFAULT 15,
    sessions_before_long_break INT DEFAULT 4,
    
    -- Default is_strict for new sessions
    default_is_strict BOOLEAN DEFAULT false,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_session_preferences_user_id 
    ON user_session_preferences(user_id);
