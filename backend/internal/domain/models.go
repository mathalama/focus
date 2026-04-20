package domain

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	Role              string    `json:"role"`
	NectarBalance     int       `json:"nectar_balance"`
	TotalNectarEarned int       `json:"total_nectar_earned"`
	CreatedAt         time.Time `json:"created_at"`
}

type AuthSession struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	UserAgent  string     `json:"user_agent"`
	IPAddress  string     `json:"ip_address"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt time.Time  `json:"last_used_at"`
}

type EmailDelivery struct {
	ID            string     `json:"id"`
	ToEmail       string     `json:"to_email"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	LastError     string     `json:"last_error,omitempty"`
	NextAttemptAt *time.Time `json:"next_attempt_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
}

type ProductEvent struct {
	ID         string         `json:"id"`
	UserID     *string        `json:"user_id,omitempty"`
	EventName  string         `json:"event_name"`
	Source     string         `json:"source"`
	Properties map[string]any `json:"properties"`
	CreatedAt  time.Time      `json:"created_at"`
}



type Goal struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	Topic              string     `json:"topic"`
	DesiredResult      string     `json:"desired_result"`
	RecommendedMinutes int        `json:"recommended_minutes"`
	Tags               []string   `json:"tags"`
	CompletedAt        *time.Time `json:"completed_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

type FocusSession struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	GoalID             string     `json:"goal_id"`
	RecommendedMinutes int        `json:"recommended_minutes"`
	IsStrict           bool       `json:"is_strict"`
	Status             string     `json:"status"` // active | paused | completed | abandoned
	PauseCount         int        `json:"pause_count"`
	StartedAt          time.Time  `json:"started_at"`
	PausedAt           *time.Time `json:"paused_at"`
	CompletedAt        *time.Time `json:"completed_at"`
}

type SessionHistoryEntry struct {
	SessionID          string     `json:"session_id"`
	GoalID             string     `json:"goal_id"`
	Topic              string     `json:"topic"`
	DesiredResult      string     `json:"desired_result"`
	Tags               []string   `json:"tags"`
	Status             string     `json:"status"` // completed | abandoned
	RecommendedMinutes int        `json:"recommended_minutes"`
	PauseCount         int        `json:"pause_count"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
}

type SessionHistorySummary struct {
	CompletedCount int     `json:"completed_count"`
	TotalMinutes   int     `json:"total_minutes"`
	AverageMinutes float64 `json:"average_minutes"`
}

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	Type        string `json:"type"`
}

type UserItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ItemID      string    `json:"item_id"`
	PurchasedAt time.Time `json:"purchased_at"`
	Item        *Item     `json:"item,omitempty"`
}

type LeaderboardEntry struct {
	UserID            string `json:"user_id"`
	Name              string `json:"name"`
	TotalNectarEarned int    `json:"total_nectar_earned"`
	Rank              int    `json:"rank"`
}

type DailyActivity struct {
	Date         string `json:"date"` // YYYY-MM-DD
	SessionCount int    `json:"session_count"`
	TotalMinutes int    `json:"total_minutes"`
}

type DailyContribution struct {
	SessionID   string     `json:"session_id"`
	GoalID      string     `json:"goal_id"`
	Topic       string     `json:"topic"`
	Minutes     int        `json:"minutes"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type Interruption struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Kind      string    `json:"kind"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type Reflection struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	WhatLearned string    `json:"what_learned"`
	WhatWasHard string    `json:"what_was_hard"`
	NextAction  string    `json:"next_action"`
	CreatedAt   time.Time `json:"created_at"`
}

type AnalyticsOverview struct {
	CompletedGoals               int     `json:"completed_goals"`
	CalmScore                    float64 `json:"calm_score"`
	FocusStability               float64 `json:"focus_stability"`
	BestHourOfDayUTC             int     `json:"best_hour_of_day_utc"`
	TotalNectarEarned            int     `json:"total_nectar_earned"`
	CalmScoreBase                float64 `json:"calm_score_base"`
	CalmScorePausePenalty        float64 `json:"calm_score_pause_penalty"`
	CalmScoreInterruptionPenalty float64 `json:"calm_score_interruption_penalty"`
	SessionsTotal                int     `json:"sessions_total"`
	CompletedSessions            int     `json:"completed_sessions"`
	PausedSessions               int     `json:"paused_sessions"`
	OtherSessions                int     `json:"other_sessions"`
	TotalPauses                  int     `json:"total_pauses"`
	TotalInterruptions           int     `json:"total_interruptions"`
	PrimaryAction                string  `json:"primary_action"`
}

type Insight struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"` // 'tip', 'warning', 'encouragement'
}

type NotificationSound struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	SessionCompleteSound string    `json:"session_complete_sound"`
	BreakEndSound        string    `json:"break_end_sound"`
	NotificationSound    string    `json:"notification_sound"`
	Volume               float64   `json:"volume"`
	SoundsEnabled        bool      `json:"sounds_enabled"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type AuthUser struct {
	User            User       `json:"-"`
	PasswordHash    string     `json:"-"`
	EmailVerifiedAt *time.Time `json:"-"`
}
type UserSessionPreferences struct {
	ID                      string    `json:"id"`
	UserID                  string    `json:"user_id"`
	PresetDurations         []int     `json:"preset_durations"` // [25, 45, 90] etc
	DefaultDuration         int       `json:"default_duration"`
	ShortBreakDuration      int       `json:"short_break_duration"`
	LongBreakDuration       int       `json:"long_break_duration"`
	SessionsBeforeLongBreak int       `json:"sessions_before_long_break"`
	DefaultIsStrict         bool      `json:"default_is_strict"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type NotificationSchedule struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	Enabled          bool       `json:"enabled"`
	Timezone         string     `json:"timezone"`
	DaysOfWeek       []int      `json:"days_of_week"`      // 0-6 (Monday-Sunday)
	ReminderTime     string     `json:"reminder_time"`     // HH:MM format
	NotificationType string     `json:"notification_type"` // start_session | daily_summary | motivational
	LastSentAt       *time.Time `json:"last_sent_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
