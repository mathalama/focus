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

// Listen for messages from popup
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'startFocus') {
    startFocusSession(request.duration, request.blockedSites).then(sendResponse);
  } else if (request.action === 'stopFocus') {
    stopFocusSession().then(sendResponse);
  } else if (request.action === 'getFocusStatus') {
    getFocusStatus().then(sendResponse);
  } else if (request.action === 'updateBlockedSites') {
    updateBlockedSites(request.sites).then(sendResponse);
  }
  return true; // Will respond asynchronously
});

async function startFocusSession(duration: number, blockedSites: BlockedSite[]): Promise<boolean> {
  const session: FocusSession = {
    active: true,
    startTime: Date.now(),
    duration: duration * 60 * 1000, // Convert to milliseconds
    blockedSites: blockedSites
  };

  await chrome.storage.local.set({ focusSession: session });
  
  // Set up timer to end session
  setTimeout(() => {
    endFocusSession();
  }, session.duration);

  // Update icon and badge
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
    iconUrl: 'assets/icon-128.png',
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

// Initialize on install/update
chrome.runtime.onInstalled.addListener(() => {
  initStorage();
});

// Keep service worker warm
chrome.alarms.create('keepAlive', { periodInMinutes: 1 });
chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === 'keepAlive') {
    // Just a ping to keep worker alive
  }
});
