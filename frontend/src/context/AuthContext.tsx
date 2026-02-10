import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { User } from '../types';
import { api } from '../api';

interface AuthContextType {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  login: (user: User, token: string, refreshToken?: string) => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
  authChecked: boolean;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [refreshToken, setRefreshToken] = useState<string | null>(() => localStorage.getItem('refresh_token'));
  const [authChecked, setAuthChecked] = useState<boolean>(() => !localStorage.getItem('token'));

  const login = useCallback((newUser: User, newToken: string, newRefreshToken?: string) => {
    localStorage.setItem('token', newToken);
    if (newRefreshToken && newRefreshToken.trim()) {
      localStorage.setItem('refresh_token', newRefreshToken);
      setRefreshToken(newRefreshToken);
    }
    localStorage.setItem('user', JSON.stringify(newUser));
    setToken(newToken);
    setUser(newUser);
    setAuthChecked(true);
  }, []);

  const logout = useCallback(() => {
    const rawRefreshToken = localStorage.getItem('refresh_token');
    if (rawRefreshToken) {
      void api.auth.logout(rawRefreshToken).catch(() => undefined);
    }

    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    setToken(null);
    setRefreshToken(null);
    setUser(null);
    setAuthChecked(true);
  }, []);

  useEffect(() => {
    let isCancelled = false;

    const validateSession = async () => {
      if (!token) {
        setUser(null);
        setAuthChecked(true);
        return;
      }

      const storedUser = localStorage.getItem('user');
      if (storedUser) {
        try {
          setUser(JSON.parse(storedUser));
        } catch {
          localStorage.removeItem('user');
          setUser(null);
        }
      } else {
        setUser(null);
      }

      try {
        const data = await api.auth.getMe();
        if (isCancelled) return;
        setUser(data.user);
        localStorage.setItem('user', JSON.stringify(data.user));
      } catch {
        if (isCancelled) return;

        if (refreshToken) {
          try {
            const refreshed = await api.auth.refresh(refreshToken);
            if (isCancelled) return;

            localStorage.setItem('token', refreshed.token);
            if (refreshed.refresh_token) {
              localStorage.setItem('refresh_token', refreshed.refresh_token);
              setRefreshToken(refreshed.refresh_token);
            }
            localStorage.setItem('user', JSON.stringify(refreshed.user));
            setToken(refreshed.token);
            setUser(refreshed.user);
            return;
          } catch {
            if (isCancelled) return;
          }
        }

        logout();
      } finally {
        if (!isCancelled) {
          setAuthChecked(true);
        }
      }
    };

    setAuthChecked(false);
    void validateSession();

    return () => {
      isCancelled = true;
    };
  }, [token, refreshToken, logout]);

  const refreshUser = useCallback(async () => {
    if (!token) return;
    try {
      const data = await api.auth.getMe();
      setUser(data.user);
      localStorage.setItem('user', JSON.stringify(data.user));
    } catch (error) {
      logout();
      throw error;
    }
  }, [token, logout]);

  return (
    <AuthContext.Provider value={{ user, token, refreshToken, login, logout, refreshUser, authChecked, isAuthenticated: !!token && !!user }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
