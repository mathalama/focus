export type SessionStatus = 'active' | 'paused' | 'completed' | 'cancelled';

export interface User {
  id: string;
  email: string;
  name: string;
  nectar_balance: number;
  total_nectar_earned: number;
  created_at: string;
}

export interface Goal {
  id: string;
  user_id: string;
  topic: string;
  desired_result: string;
  recommended_minutes: number;
  tags: string[];
  created_at: string;
}

export interface FocusSession {
  id: string;
  user_id: string;
  goal_id: string;
  recommended_minutes: number;
  is_strict: boolean;
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

export interface Interruption {
  id: string;
  session_id: string;
  kind: string;
  reason: string;
  created_at: string;
}

export interface AnalyticsOverview {
  completed_goals: number;
  calm_score: number;
  focus_stability: number;
  best_hour_of_day_utc: number;
  total_nectar_earned: number;
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
  user_id: string;
  item_id: string;
  purchased_at: string;
}

export interface DailyActivity {
  date: string;
  session_count: number;
  total_minutes: number;
}

export interface Insight {
  title: string;
  content: string;
  type: 'tip' | 'warning' | 'encouragement';
}
