// Browser push notifications

export interface NotificationOption {
  title: string;
  body?: string;
  icon?: string;
  badge?: string;
  tag?: string;
  requireInteraction?: boolean;
  vibrate?: number | number[];
}

/**
 * Request permission for browser notifications (if not already granted)
 */
export const requestNotificationPermission = async (): Promise<boolean> => {
  if (typeof window === 'undefined') return false;
  if (!('Notification' in window)) {
    console.log('Browser does not support notifications');
    return false;
  }

  if (Notification.permission === 'granted') {
    return true;
  }

  if (Notification.permission !== 'denied') {
    try {
      const permission = await Notification.requestPermission();
      return permission === 'granted';
    } catch (err) {
      console.warn('Failed to request notification permission:', err);
      return false;
    }
  }

  return false;
};

/**
 * Check if notifications are enabled
 */
export const notificationsEnabled = (): boolean => {
  if (typeof window === 'undefined') return false;
  return 'Notification' in window && Notification.permission === 'granted';
};

/**
 * Show a browser notification
 */
export const showNotification = async (options: NotificationOption): Promise<void> => {
  if (typeof window === 'undefined') return;
  if (!('Notification' in window)) return;

  // Check if notifications are disabled in localStorage
  if (localStorage.getItem('notifications-enabled') === 'false') return;

  if (Notification.permission !== 'granted') {
    const granted = await requestNotificationPermission();
    if (!granted) return;
  }

  try {
    const notification = new Notification(options.title, {
      body: options.body,
      icon: options.icon || '/logo.svg',
      badge: options.badge || '/logo.svg',
      tag: options.tag,
      requireInteraction: options.requireInteraction ?? false,
    });

    // Auto close after 5 seconds (unless requireInteraction is true)
    if (!options.requireInteraction) {
      setTimeout(() => notification.close(), 5000);
    }

    // Handle click
    notification.onclick = () => {
      window.focus();
      notification.close();
    };
  } catch (err) {
    console.warn('Failed to show notification:', err);
  }
};

/**
 * Show session completion notification
 */
export const notifySessionComplete = async (): Promise<void> => {
  await showNotification({
    title: '🎉 Сессия завершена!',
    body: 'Отлично поработали! Время для отражения.',
    tag: 'session-complete',
    requireInteraction: false,
  });
};

/**
 * Show break end notification
 */
export const notifyBreakEnd = async (): Promise<void> => {
  await showNotification({
    title: '⏰ Перерыв закончился',
    body: 'Готовы продолжить?',
    tag: 'break-end',
    requireInteraction: false,
  });
};
