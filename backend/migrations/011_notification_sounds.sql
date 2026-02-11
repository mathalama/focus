-- Create notification sounds table
CREATE TABLE notification_sounds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Sound URLs/references
  session_complete_sound TEXT DEFAULT '/sounds/session-complete.mp3',
  break_end_sound TEXT DEFAULT '/sounds/break-end.mp3',
  notification_sound TEXT DEFAULT '/sounds/notification.mp3',
  
  -- Sound volume (0.0 - 1.0)
  volume NUMERIC(3, 2) DEFAULT 0.7,
  
  -- Flags
  sounds_enabled BOOLEAN DEFAULT true,
  
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  
  UNIQUE(user_id)
);

-- Create index for user lookups
CREATE INDEX idx_notification_sounds_user_id ON notification_sounds(user_id);

-- Initialize notification_sounds for existing users
INSERT INTO notification_sounds (user_id)
SELECT id FROM users
ON CONFLICT (user_id) DO NOTHING;
