import { authAPI } from './auth';
import { sessionsAPI } from './sessions';
import { goalsAPI } from './goals';
import { analyticsAPI } from './analytics';
import { gamificationAPI } from './gamification';
import { adminAPI } from './admin';

export const api = {
  auth: authAPI,
  sessions: sessionsAPI,
  goals: goalsAPI,
  analytics: analyticsAPI,
  gamification: gamificationAPI,
  admin: adminAPI,
};
