export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

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
  topic: string;
  desired_result: string;
  recommended_minutes: number;
  tags?: string[];
  completed_at?: string;
  created_at: string;
}

export interface FocusSession {
  id: string;
  goal_id: string;
  status: 'active' | 'paused' | 'completed' | 'cancelled';
  recommended_minutes: number;
  is_strict: boolean;
  pause_count: number;
  started_at: string;
  paused_at?: string;
  completed_at?: string;
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

export interface AnalyticsOverview {
  completed_goals: number;
  calm_score: number;
  focus_stability: number;
  best_hour_of_day_utc: number;
  total_nectar_earned: number;
}

export interface Insight {
  title: string;
  content: string;
  type: 'tip' | 'warning' | 'encouragement';
}

const getAuthHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { 'Authorization': `Bearer ${token}` } : {};
};

export const api = {
  auth: {
    devLogin: async (email: string, name: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/auth/dev-login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, name }),
      });
      if (!res.ok) throw new Error('Login failed');
      return res.json();
    },
    getMe: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/me`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get user');
      return res.json();
    },
  },
  goals: {
    create: async (data: { topic: string; desired_result: string; recommended_minutes: number; tags: string[] }) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/goals`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error('Failed to create goal');
      return res.json();
    },
    list: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/goals`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to list goals');
      return res.json();
    },
    history: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/goals/history`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to list goal history');
      return res.json();
    },
  },
  sessions: {
    start: async (data: { goal_id: string; recommended_minutes: number; is_strict: boolean }) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error('Failed to start session');
      return res.json();
    },
    get: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get session');
      return res.json();
    },
    pause: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/pause`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to pause session');
      return res.json();
    },
    resume: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/resume`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to resume session');
      return res.json();
    },
    reset: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/reset`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to reset session');
      return res.json();
    },
    abandon: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/abandon`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to abandon session');
      return res.json();
    },
    complete: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/complete`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to complete session');
      return res.json();
    },
    addInterruption: async (sessionID: string, reason: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/interruption`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify({ reason }),
      });
      if (!res.ok) throw new Error('Failed to add interruption');
      return res.json();
    },
    reflection: async (sessionID: string, data: { what_learned: string; what_was_hard: string; next_action: string }) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/reflection`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error('Failed to submit reflection');
      return res.json();
    },
  },
  analytics: {
    overview: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/analytics/overview`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get analytics');
      return res.json();
    },
    activity: async (timezone?: string) => {
      const url = new URL(`${API_BASE_URL}/api/v1/analytics/activity`);
      if (timezone) url.searchParams.append('timezone', timezone);
      const res = await fetch(url.toString(), {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get activity');
      return res.json();
    },
    insights: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/analytics/insights`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get insights');
      return res.json();
    }
  },
  gamification: {
    leaderboard: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/leaderboard`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get leaderboard');
      return res.json();
    },
    items: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/shop/items`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get items');
      return res.json();
    },
    buy: async (itemID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/shop/items/${itemID}/buy`, {
        method: 'POST',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to buy item');
      return res.json();
    }
  }
};
