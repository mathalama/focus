export type SessionStatus = 'active' | 'paused' | 'completed' | 'cancelled';

export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface Goal {
  id: string;
  user_id: string;
  topic: string;
  desired_result: string;
  recommended_minutes: number;
  created_at: string;
}

export interface FocusSession {
  id: string;
  user_id: string;
  goal_id: string;
  recommended_minutes: number;
  status: SessionStatus;
  pause_count: number;
  started_at: string;
  paused_at: string | null;
  completed_at: string | null;
}

export interface Reflection {
  id: string;
  session_id: string;
  what_learned: string;
  what_was_hard: string;
  next_action: string;
  created_at: string;
}

export interface AnalyticsOverview {
  completed_goals: number;
  calm_score: number;
  focus_stability: number;
  best_hour_of_day_utc: number;
}
