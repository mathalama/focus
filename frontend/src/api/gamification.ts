import { API_BASE_URL, getAuthHeaders } from './client';

export const gamificationAPI = {
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
  },
};
