import type { SessionHistoryEntry, SessionHistoryFilters, SessionHistorySummary } from '../types';
import { API_BASE_URL, createIdempotencyKey, getAuthHeaders, normalizeSessionResponse } from './client';

export const sessionsAPI = {
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
      headers: { ...getAuthHeaders(), 'Idempotency-Key': createIdempotencyKey() },
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
};
