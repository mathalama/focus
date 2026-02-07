import type { AnalyticsOverview, FocusSession, Goal, Reflection, User } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

type Method = 'GET' | 'POST' | 'PATCH';

interface ApiOptions {
  method?: Method;
  body?: unknown;
  userID?: string;
}

async function request<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };

  if (options.userID) {
    headers['X-User-ID'] = options.userID;
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

export async function devLogin(email: string, name: string): Promise<User> {
  const response = await request<{ user: User }>('/api/v1/auth/dev-login', {
    method: 'POST',
    body: { email, name }
  });
  return response.user;
}

export async function listGoals(userID: string): Promise<Goal[]> {
  const response = await request<{ goals: Goal[] }>('/api/v1/goals', { userID });
  return response.goals;
}

export async function createGoal(
  userID: string,
  payload: { topic: string; desired_result: string; recommended_minutes: number }
): Promise<Goal> {
  const response = await request<{ goal: Goal }>('/api/v1/goals', {
    method: 'POST',
    body: payload,
    userID
  });
  return response.goal;
}

export async function startSession(
  userID: string,
  payload: { goal_id: string; recommended_minutes: number }
): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>('/api/v1/sessions', {
    method: 'POST',
    body: payload,
    userID
  });
  return response.session;
}

export async function pauseSession(userID: string, sessionID: string): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>(`/api/v1/sessions/${sessionID}/pause`, {
    method: 'PATCH',
    userID
  });
  return response.session;
}

export async function resumeSession(userID: string, sessionID: string): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>(`/api/v1/sessions/${sessionID}/resume`, {
    method: 'PATCH',
    userID
  });
  return response.session;
}

export async function completeSession(userID: string, sessionID: string): Promise<FocusSession> {
  const response = await request<{ session: FocusSession }>(`/api/v1/sessions/${sessionID}/complete`, {
    method: 'PATCH',
    userID
  });
  return response.session;
}

export async function addInterruption(userID: string, sessionID: string, reason: string): Promise<void> {
  await request(`/api/v1/sessions/${sessionID}/interruption`, {
    method: 'POST',
    userID,
    body: { reason }
  });
}

export async function saveReflection(
  userID: string,
  sessionID: string,
  payload: { what_learned: string; what_was_hard: string; next_action: string }
): Promise<Reflection> {
  const response = await request<{ reflection: Reflection }>(`/api/v1/sessions/${sessionID}/reflection`, {
    method: 'POST',
    userID,
    body: payload
  });
  return response.reflection;
}

export async function getOverview(userID: string): Promise<AnalyticsOverview> {
  const response = await request<{ overview: AnalyticsOverview }>('/api/v1/analytics/overview', {
    userID
  });
  return response.overview;
}
