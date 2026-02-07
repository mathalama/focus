package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Goal struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	Topic              string    `json:"topic"`
	DesiredResult      string    `json:"desired_result"`
	RecommendedMinutes int       `json:"recommended_minutes"`
	CreatedAt          time.Time `json:"created_at"`
}

type FocusSession struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	GoalID             string     `json:"goal_id"`
	RecommendedMinutes int        `json:"recommended_minutes"`
	Status             string     `json:"status"`
	PauseCount         int        `json:"pause_count"`
	StartedAt          time.Time  `json:"started_at"`
	PausedAt           *time.Time `json:"paused_at"`
	CompletedAt        *time.Time `json:"completed_at"`
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
	CompletedGoals   int     `json:"completed_goals"`
	CalmScore        float64 `json:"calm_score"`
	FocusStability   float64 `json:"focus_stability"`
	BestHourOfDayUTC int     `json:"best_hour_of_day_utc"`
}
