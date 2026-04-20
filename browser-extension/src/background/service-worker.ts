// Background service worker for managing focus mode state and rules
interface BlockedSite {
  domain: string;
  pattern: string;
  enabled: boolean;
}

interface FocusSession {
  active: boolean;
  startTime: number;
  duration: number;
  blockedSites: BlockedSite[];
}

// Initialize storage
async function initStorage() {
  const data = await chrome.storage.local.get(['focusSession', 'defaultBlockedSites']);
  
  if (!data.focusSession) {
    await chrome.storage.local.set({
      focusSession: {
        active: false,
        startTime: 0,
        duration: 0,
        blockedSites: []
      }
    });
  }

  if (!data.defaultBlockedSites) {
    await chrome.storage.local.set({
      defaultBlockedSites: [
        { domain: 'youtube.com', pattern: '*://*.youtube.com/*', enabled: true },
        { domain: 'reddit.com', pattern: '*://*.reddit.com/*', enabled: true },
        { domain: 'twitter.com', pattern: '*://*.twitter.com/*', enabled: true },
        { domain: 'facebook.com', pattern: '*://*.facebook.com/*', enabled: true },
        { domain: 'instagram.com', pattern: '*://*.instagram.com/*', enabled: true },
        { domain: 'tiktok.com', pattern: '*://*.tiktok.com/*', enabled: true },
        { domain: 'twitch.tv', pattern: '*://*.twitch.tv/*', enabled: true }
      ]
    });
  }
}

const API_BASE_URL = 'http://localhost:8080';

// Backend communication service
const BackendService = {
  async getToken(): Promise<string | null> {
    const data = await chrome.storage.local.get('authToken');
    return data.authToken || null;
  },

  async setToken(token: string): Promise<void> {
    await chrome.storage.local.set({ authToken: token });
  },

  async sendFocusEvent(action: 'start' | 'stop', durationMinutes?: number): Promise<boolean> {
    const token = await this.getToken();
    if (!token) {
      console.warn('Cannot send focus event: no auth token');
      return false;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/focus/event`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          action,
          duration: durationMinutes || 0
        })
      });

      if (!response.ok) {
        console.error('Failed to send focus event:', response.statusText);
        return false;
      }

      return true;
    } catch (error) {
      console.error('Error sending focus event:', error);
      return false;
    }
  }
};

// Listen for messages from content script or popup
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'getFocusStatus') {
    getFocusStatus().then(sendResponse);
  } else if (request.action === 'focusStateChanged') {
    // Focus state changed from website via content script
    console.log('Focus state changed:', request.state);
    if (request.state === 'active') {
      BackendService.sendFocusEvent('start', request.duration);
    } else if (request.state === 'completed' || request.state === 'abandoned') {
      BackendService.sendFocusEvent('stop');
    }
    sendResponse({ ok: true });
  } else if (request.action === 'stopFocus') {
    // User clicked End Focus button in extension
    stopFocusSession().then((ok) => {
      BackendService.sendFocusEvent('stop');
      sendResponse(ok);
    });
  } else if (request.action === 'setToken') {
    console.log('Token received from content script');
    BackendService.setToken(request.token).then(() => {
      sendResponse({ ok: true });
    });
  }
  return true; // Will respond asynchronously
});

async function startFocusSession(duration: number, blockedSites: BlockedSite[]): Promise<boolean> {
  const session: FocusSession = {
    active: true,
    startTime: Date.now(),
    duration: duration * 60 * 1000,
    blockedSites: blockedSites
  };

  await chrome.storage.local.set({ focusSession: session });
  
  setTimeout(() => {
    endFocusSession();
  }, session.duration);

  await updateIcon(true, duration);

  return true;
}

async function stopFocusSession(): Promise<boolean> {
  const session: FocusSession = {
    active: false,
    startTime: 0,
    duration: 0,
    blockedSites: []
  };

  await chrome.storage.local.set({ focusSession: session });
  await updateIcon(false, 0);

  return true;
}

async function endFocusSession(): Promise<void> {
  await stopFocusSession();
  
  // Show notification that focus session ended
  chrome.notifications.create('focusComplete', {
    type: 'basic',
    iconUrl: 'assets/icon-48.svg',
    title: 'Focus Session Complete!',
    message: 'Great work on staying focused!'
  });
}

async function getFocusStatus(): Promise<{ focus: FocusSession; elapsed: number }> {
  const data = await chrome.storage.local.get('focusSession');
  const session = data.focusSession as FocusSession;
  
  let elapsed = 0;
  if (session.active) {
    elapsed = Date.now() - session.startTime;
  }

  return { focus: session, elapsed };
}

async function updateBlockedSites(sites: BlockedSite[]): Promise<boolean> {
  const data = await chrome.storage.local.get('focusSession');
  const session = data.focusSession as FocusSession;
  
  session.blockedSites = sites;
  await chrome.storage.local.set({ focusSession: session });

  return true;
}

async function updateIcon(active: boolean, remainingMinutes: number): Promise<void> {
  if (active) {
    chrome.action.setBadgeBackgroundColor({ color: '#FF6B6B' });
    chrome.action.setBadgeText({ text: remainingMinutes.toString() });
  } else {
    chrome.action.setBadgeText({ text: '' });
  }
}

// Listen for storage changes
chrome.storage.onChanged.addListener((changes, area) => {
  if (area !== 'local' || !changes.focusSession) return;

  const newSession = changes.focusSession.newValue as FocusSession | undefined;
  
  if (newSession?.active) {
    const remainingMinutes = Math.ceil(newSession.duration / 60000);
    updateIcon(true, remainingMinutes);
    
    // Schedule end of session
    setTimeout(() => {
      endFocusSession();
    }, newSession.duration);
  } else {
    updateIcon(false, 0);
  }
});

// Initialize on install/update
chrome.runtime.onInstalled.addListener(() => {
  initStorage();
});
