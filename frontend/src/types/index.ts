export interface User {
  id: string;
  email: string;
  name: string;
  role: 'user' | 'admin';
  nectar_balance: number;
  total_nectar_earned: number;
  created_at: string;
}

export interface AuthSession {
  id: string;
  user_id: string;
  user_agent: string;
  ip_address: string;
  expires_at: string;
  revoked_at?: string;
  created_at: string;
  last_used_at: string;
}

export interface Goal {
  id: string;
  topic: string;
  desired_result: string;
  recommended_minutes: number;
  tags?: string[];
  completed_at?: string;
  created_at: string;
}

export type SessionStatus = 'active' | 'paused' | 'completed' | 'abandoned';

export interface FocusSession {
  id: string;
  goal_id: string;
  status: SessionStatus;
  recommended_minutes: number;
  is_strict: boolean;
  pause_count: number;
  started_at: string;
  paused_at?: string;
  completed_at?: string;
}

export interface SessionHistoryEntry {
  session_id: string;
  goal_id: string;
  topic: string;
  desired_result: string;
  tags: string[];
  status: SessionStatus;
  recommended_minutes: number;
  pause_count: number;
  started_at: string;
  completed_at?: string;
}

export interface SessionHistorySummary {
  completed_count: number;
  total_minutes: number;
  average_minutes: number;
}

export interface SessionHistoryFilters {
  period?: 'today' | 'week' | 'month' | 'all';
  tags?: string[];
  min_minutes?: number;
  max_minutes?: number;
  status?: 'completed' | 'abandoned' | 'all';
  timezone?: string;
  limit?: number;
}

export interface LeaderboardEntry {
  user_id: string;
  name: string;
  total_nectar_earned: number;
  rank: number;
}

export interface ShopItem {
  id: string;
  name: string;
  description: string;
  cost: number;
  type: string;
}

export interface UserItem {
  id: string;
  item_id: string;
  purchased_at: string;
}

export interface Interruption {
  id: string;
  session_id: string;
  kind: string;
  reason: string;
  created_at: string;
}

export interface DailyActivity {
  date: string;
  session_count: number;
  total_minutes: number;
}

export interface DailyContribution {
  session_id: string;
  goal_id: string;
  topic: string;
  minutes: number;
  started_at: string;
  completed_at?: string;
}

export interface AnalyticsOverview {
  completed_goals: number;
  calm_score: number;
  focus_stability: number;
  calm_score_base: number;
  calm_score_pause_penalty: number;
  calm_score_interruption_penalty: number;
  sessions_total: number;
  completed_sessions: number;
  paused_sessions: number;
  other_sessions: number;
  total_pauses: number;
  total_interruptions: number;
  primary_action: 'start_sessions' | 'complete_more_sessions' | 'reduce_pauses' | 'reduce_interruptions' | 'stabilize_schedule' | 'keep_momentum';
  best_hour_of_day_utc: number;
  total_nectar_earned: number;
}

export interface Insight {
  title: string;
  content: string;
  type: 'tip' | 'warning' | 'encouragement';
}

export interface AuthResponse {
  user: User;
  token: string;
  refresh_token?: string;
}
