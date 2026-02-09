import type { FocusSession, SessionStatus } from '../types';

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const getAuthHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('token');
  return token ? { 'Authorization': `Bearer ${token}` } : {};
};

export const normalizeSession = (session: any): FocusSession => {
  const rawStatus = session?.status;
  const status: SessionStatus = rawStatus === 'cancelled' ? 'abandoned' : rawStatus;
  return { ...session, status };
};

export const normalizeSessionResponse = <T extends { session?: any | null }>(data: T): T => {
  if (!data?.session) return data;
  return { ...data, session: normalizeSession(data.session) };
};

export const readAPIError = async (res: Response, fallback: string): Promise<string> => {
  try {
    const payload = await res.json();
    if (typeof payload?.error === 'string' && payload.error.trim()) {
      return payload.error.trim();
    }
  } catch {
    // Keep fallback message for non-JSON errors.
  }
  return fallback;
};
