import { API_BASE_URL, getAuthHeaders } from './client';

export interface NotificationSound {
  id: string;
  user_id: string;
  session_complete_sound: string;
  break_end_sound: string;
  notification_sound: string;
  volume: number;
  sounds_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export const notifications = {
  getSounds: async (): Promise<{ sound: NotificationSound }> => {
    const res = await fetch(`${API_BASE_URL}/api/v1/notifications/sounds`, {
      headers: getAuthHeaders(),
    });
    if (!res.ok) throw new Error('Failed to get notification sounds');
    return res.json();
  },

  updateSounds: async (data: Partial<Omit<NotificationSound, 'id' | 'user_id' | 'created_at' | 'updated_at'>>) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/notifications/sounds`, {
      method: 'PATCH',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error('Failed to update notification sounds');
    return res.json();
  },
};
