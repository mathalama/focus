import { API_BASE_URL, getAuthHeaders } from './client';

export const analyticsAPI = {
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
  },
};
