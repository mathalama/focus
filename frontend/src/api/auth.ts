import type { AuthResponse } from '../types';
import { API_BASE_URL, getAuthHeaders, readAPIError } from './client';

export const authAPI = {
  register: async (email: string, name: string, password: string) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, name, password }),
    });
    if (!res.ok) throw new Error(await readAPIError(res, 'Registration failed'));
    return res.json();
  },

  login: async (email: string, password: string): Promise<AuthResponse> => {
    const res = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) throw new Error(await readAPIError(res, 'Login failed'));
    return res.json();
  },

  resendVerification: async (email: string) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/auth/verify-email/resend`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    });
    if (!res.ok) throw new Error(await readAPIError(res, 'Failed to resend verification email'));
    return res.json();
  },

  devLogin: async (email: string, name: string): Promise<AuthResponse> => {
    const res = await fetch(`${API_BASE_URL}/api/v1/auth/dev-login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, name }),
    });
    if (!res.ok) throw new Error(await readAPIError(res, 'Login failed'));
    return res.json();
  },

  getMe: async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/me`, {
      headers: getAuthHeaders(),
    });
    if (!res.ok) throw new Error('Failed to get user');
    return res.json();
  },
};
