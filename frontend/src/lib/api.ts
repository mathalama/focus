import type { AnalyticsOverview, FocusSession, Goal, Reflection, User } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

type Method = 'GET' | 'POST' | 'PATCH';

interface ApiOptions {
  method?: Method;
  body?: unknown;
  token?: string;
}

async function request<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };

  if (options.token) {
    headers['Authorization'] = `Bearer ${options.token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method: options.method ?? 'GET',
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok) {
    const errorMessage = (payload as { error?: string }).error ?? 'Request failed';
    throw new Error(errorMessage);
  }

  return payload as T;
}

export async function devLogin(email: string, name: string): Promise<{ user: User; token: string }> {
  const response = await request<{ user: User; token: string }>('/api/v1/auth/dev-login', {
    method: 'POST',
    body: { email, name }
  });
  return response;
}

export async function listGoals(token: string): Promise<Goal[]> {
  const response = await request<{ goals: Goal[] }>('/api/v1/goals', { token });
  return response.goals;
}

export async function createGoal(
  token: string,
  payload: { topic: string; desired_result: string; recommended_minutes: number }
): Promise<Goal> {
  const response = await request<{ goal: Goal }>('/api/v1/goals', {
    method: 'POST',
    body: payload,
    token
  });
  return response.goal;
}

export async function startSession(
  token: string,
  payload: { goal_id: string; recommended_minutes: number }
): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>('/api/v1/sessions', {
    method: 'POST',
    body: payload,
    token
  });
  return response.session;
}

export async function pauseSession(token: string, sessionID: string): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>(`/api/v1/sessions/${sessionID}/pause`, {
    method: 'PATCH',
    token
  });
  return response.session;
}

export async function resumeSession(token: string, sessionID: string): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>(`/api/v1/sessions/${sessionID}/resume`, {
    method: 'PATCH',
    token
  });
  return response.session;
}

export async function completeSession(token: string, sessionID: string): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>(`/api/v1/sessions/${sessionID}/complete`, {
    method: 'PATCH',
    token
  });
  return response.session;
}

export async function addInterruption(token: string, sessionID: string, reason: string): Promise<void> {
  await request(`/api/v1/sessions/${sessionID}/interruption`, {
    method: 'POST',
    token,
    body: { reason }
  });
}

export async function saveReflection(
  token: string,
  sessionID: string,
  payload: { what_learned: string; what_was_hard: string; next_action: string }
): Promise<Reflection> {
  const response = await request<{ reflection: Reflection }>(`/api/v1/sessions/${sessionID}/reflection`, {
    method: 'POST',
    token,
    body: payload
  });
  return response.reflection;
}

export async function getOverview(token: string): Promise<AnalyticsOverview> {
  const response = await request<{ overview: AnalyticsOverview }>('/api/v1/analytics/overview', {
    token
  });
  return response.overview;
}
