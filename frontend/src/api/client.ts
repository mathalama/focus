import type { FocusSession, SessionStatus } from '../types';

const DEV_API_BASE_URL = 'http://localhost:8080';

const normalizeBaseURL = (value: string): string => value.trim().replace(/\/+$/, '');

const isLocalAddress = (value: string): boolean => {
  return /^https?:\/\/(localhost|127(?:\.\d{1,3}){3})(?::\d+)?$/i.test(value);
};

const envBaseURL = typeof import.meta.env.VITE_API_BASE_URL === 'string'
  ? normalizeBaseURL(import.meta.env.VITE_API_BASE_URL)
  : '';

if (import.meta.env.PROD) {
  if (!envBaseURL) {
    throw new Error('Missing VITE_API_BASE_URL in production build');
  }
  if (isLocalAddress(envBaseURL)) {
    throw new Error('Invalid VITE_API_BASE_URL in production build: localhost is not allowed');
  }
}

export const API_BASE_URL = envBaseURL || DEV_API_BASE_URL;

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

export const createIdempotencyKey = (): string => {
  const randomPart = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random().toString(36).slice(2)}`;
  return `idem-${randomPart}`;
};
