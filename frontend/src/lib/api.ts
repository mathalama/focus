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
  best_hour_of_day_utc: number;
  total_nectar_earned: number;
}

export interface Insight {
  title: string;
  content: string;
  type: 'tip' | 'warning' | 'encouragement';
}

export interface TelegramLinkCode {
  code: string;
  expires_at: string;
}

const getAuthHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { 'Authorization': `Bearer ${token}` } : {};
};

const normalizeSession = (session: any): FocusSession => {
  const rawStatus = session?.status;
  const status: SessionStatus = rawStatus === 'cancelled' ? 'abandoned' : rawStatus;
  return {
    ...session,
    status,
  };
};

const normalizeSessionResponse = <T extends { session?: any | null }>(data: T): T => {
  if (!data?.session) return data;
  return {
    ...data,
    session: normalizeSession(data.session),
  };
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
    createTelegramLinkCode: async (): Promise<TelegramLinkCode> => {
      const res = await fetch(`${API_BASE_URL}/api/v1/auth/telegram/link-code`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...getAuthHeaders(),
        },
      });

      if (!res.ok) {
        let message = 'Failed to generate Telegram code';
        try {
          const payload = await res.json();
          if (typeof payload?.error === 'string' && payload.error.trim()) {
            message = payload.error.trim();
          }
        } catch {
          // Ignore non-JSON errors and keep fallback message.
        }
        throw new Error(message);
      }

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
      return normalizeSessionResponse(await res.json());
    },
    active: async () => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/active`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get active session');
      return normalizeSessionResponse(await res.json());
    },
    get: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}`, {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get session');
      return normalizeSessionResponse(await res.json());
    },
    pause: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/pause`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to pause session');
      return normalizeSessionResponse(await res.json());
    },
    resume: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/resume`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to resume session');
      return normalizeSessionResponse(await res.json());
    },
    reset: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/reset`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to reset session');
      return normalizeSessionResponse(await res.json());
    },
    abandon: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/abandon`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to abandon session');
      return normalizeSessionResponse(await res.json());
    },
    complete: async (sessionID: string) => {
      const res = await fetch(`${API_BASE_URL}/api/v1/sessions/${sessionID}/complete`, {
        method: 'PATCH',
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to complete session');
      return normalizeSessionResponse(await res.json());
    },
    history: async (filters?: SessionHistoryFilters) => {
      const url = new URL(`${API_BASE_URL}/api/v1/sessions/history`);
      if (filters?.period) url.searchParams.set('period', filters.period);
      if (filters?.status) url.searchParams.set('status', filters.status);
      if (filters?.timezone) url.searchParams.set('timezone', filters.timezone);
      if (filters?.min_minutes && filters.min_minutes > 0) url.searchParams.set('min_minutes', String(filters.min_minutes));
      if (filters?.max_minutes && filters.max_minutes > 0) url.searchParams.set('max_minutes', String(filters.max_minutes));
      if (filters?.limit && filters.limit > 0) url.searchParams.set('limit', String(filters.limit));
      if (filters?.tags && filters.tags.length > 0) {
        const tagString = filters.tags.map((tag) => tag.trim()).filter(Boolean).join(',');
        if (tagString) {
          url.searchParams.set('tags', tagString);
        }
      }

      const res = await fetch(url.toString(), {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get session history');
      const data = await res.json();
      const sessions: SessionHistoryEntry[] = Array.isArray(data?.sessions)
        ? data.sessions.map((item: any) => ({
            ...item,
            status: item?.status === 'cancelled' ? 'abandoned' : item?.status,
          }))
        : [];
      return {
        sessions,
        summary: data?.summary as SessionHistorySummary | undefined,
      };
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
    day: async (date: string, timezone?: string) => {
      const url = new URL(`${API_BASE_URL}/api/v1/analytics/activity/day`);
      url.searchParams.append('date', date);
      if (timezone) url.searchParams.append('timezone', timezone);
      const res = await fetch(url.toString(), {
        headers: getAuthHeaders(),
      });
      if (!res.ok) throw new Error('Failed to get daily contributions');
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
