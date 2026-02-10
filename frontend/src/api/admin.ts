import { API_BASE_URL, getAuthHeaders } from './client';

export const adminAPI = {
  health: async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/admin/health`, {
      headers: getAuthHeaders(),
    });
    if (!res.ok) throw new Error('Failed to load admin health');
    return res.json();
  },

  events: async (limit = 50) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/admin/events?limit=${encodeURIComponent(String(limit))}`, {
      headers: getAuthHeaders(),
    });
    if (!res.ok) throw new Error('Failed to load product events');
    return res.json();
  },

  emailDeliveries: async (limit = 50) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/admin/email-deliveries?limit=${encodeURIComponent(String(limit))}`, {
      headers: getAuthHeaders(),
    });
    if (!res.ok) throw new Error('Failed to load email deliveries');
    return res.json();
  },
};
