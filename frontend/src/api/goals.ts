import { API_BASE_URL, getAuthHeaders } from './client';

export const goalsAPI = {
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
};
