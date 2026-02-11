// Content script that runs on all pages to enforce blocking

interface FocusSession {
  active: boolean;
  startTime: number;
  duration: number;
  blockedSites: Array<{ domain: string; pattern: string; enabled: boolean }>;
}

// Listen for focus state changes from the website via localStorage
window.addEventListener('storage', (event) => {
  if (event.key === 'focus-session-state' && event.newValue) {
    try {
      const data = JSON.parse(event.newValue);
      const { state } = data;
      
      console.log('Content script received focus state change:', state);
      
      // Get blocked sites from extension storage
      chrome.storage.local.get('defaultBlockedSites', (data) => {
        const blockedSites = (data.defaultBlockedSites || []).filter(
          (site: any) => site.enabled
        );
        
        // Update focus session state
        if (state === 'active') {
          // Get duration from sessionStorage if available (defaults to 25 min)
          const duration = JSON.parse(sessionStorage.getItem('focus-session-info') || '{}').duration || 25;
          
          chrome.storage.local.set({
            focusSession: {
              active: true,
              startTime: Date.now(),
              duration: duration * 60 * 1000,
              blockedSites: blockedSites
            }
          });
        } else if (state === 'paused' || state === 'completed') {
          chrome.storage.local.set({
            focusSession: {
              active: false,
              startTime: 0,
              duration: 0,
              blockedSites: []
            }
          });
        }
      });
    } catch (e) {
      console.error('Error parsing focus state:', e);
    }
  }
});

// Also listen for custom events dispatched from page
window.addEventListener('focus-session-changed', () => {
  try {
    const data = JSON.parse(localStorage.getItem('focus-session-state') || '{}');
    const { state } = data;
    
    console.log('Content script received focus state change (via custom event):', state);
    
    // Get blocked sites from extension storage
    chrome.storage.local.get('defaultBlockedSites', (data) => {
      const blockedSites = (data.defaultBlockedSites || []).filter(
        (site: any) => site.enabled
      );
      
      // Update focus session state
      if (state === 'active') {
        const duration = JSON.parse(sessionStorage.getItem('focus-session-info') || '{}').duration || 25;
        
        chrome.storage.local.set({
          focusSession: {
            active: true,
            startTime: Date.now(),
            duration: duration * 60 * 1000,
            blockedSites: blockedSites
          }
        });
      } else if (state === 'paused' || state === 'completed') {
        chrome.storage.local.set({
          focusSession: {
            active: false,
            startTime: 0,
            duration: 0,
            blockedSites: []
          }
        });
      }
    });
  } catch (e) {
    console.error('Error in focus-session-changed handler:', e);
  }
});

async function checkAndBlock(): Promise<void> {
  const data = await chrome.storage.local.get('focusSession');
  const session = data.focusSession as FocusSession;

  if (!session.active) {
    removeBlockingPage();
    return;
  }

  const currentHost = window.location.hostname;
  const isBlocked = session.blockedSites.some(
    (site) => site.enabled && isHostMatching(currentHost, site.domain)
  );

  if (isBlocked) {
    showBlockingPage(session);
  } else {
    removeBlockingPage();
  }
}

function isHostMatching(currentHost: string, domain: string): boolean {
  return (
    currentHost === domain ||
    currentHost.endsWith('.' + domain) ||
    domain.endsWith('.' + currentHost)
  );
}

function showBlockingPage(session: FocusSession): void {
  if (document.getElementById('mathalama-focus-blocker')) {
    return;
  }

  const blocker = document.createElement('div');
  blocker.id = 'mathalama-focus-blocker';
  blocker.innerHTML = `
    <style>
      #mathalama-focus-blocker {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: #1a1a1a;
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 999999;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      }
      
      .blocker-content {
        text-align: center;
        color: #e0e0e0;
        padding: 40px;
        border-radius: 12px;
        background: #2a2a2a;
        border: 1px solid #444;
        max-width: 500px;
      }
      
      .blocker-title {
        font-size: 36px;
        font-weight: 700;
        margin-bottom: 16px;
        color: #ffffff;
      }
      
      .blocker-message {
        font-size: 16px;
        margin-bottom: 24px;
        color: #b0b0b0;
      }
      
      .blocker-timer {
        font-size: 56px;
        font-weight: 300;
        margin: 24px 0;
        font-family: monospace;
        color: #ffffff;
      }
      
      .blocker-remaining {
        font-size: 13px;
        margin-bottom: 32px;
        color: #888;
      }
      
      .blocker-button {
        background: #333;
        color: #e0e0e0;
        border: 1px solid #444;
        padding: 10px 28px;
        border-radius: 6px;
        font-size: 14px;
        cursor: pointer;
        transition: all 0.3s ease;
        margin: 0 8px;
      }
      
      .blocker-button:hover {
        background: #404040;
        border-color: #555;
      }
      
      .blocker-button.danger {
        background: #4a2a2a;
        border-color: #666;
        color: #e8b8b8;
      }
      
      .blocker-button.danger:hover {
        background: #552a2a;
        border-color: #777;
      }
    </style>
    
    <div class="blocker-content">
      <div class="blocker-title">Focus Mode</div>
      <div class="blocker-message">Stay focused, avoid distractions</div>
      <div class="blocker-timer" id="timer">25:00</div>
      <div class="blocker-remaining" id="remaining">Time remaining in focus session</div>
      <div>
        <button class="blocker-button" onclick="window.history.back()">Go Back</button>
        <button class="blocker-button danger" id="end-focus">End Focus</button>
      </div>
    </div>
  `;

  document.documentElement.appendChild(blocker);
  updateTimer(session);

  document.getElementById('end-focus')?.addEventListener('click', () => {
    chrome.runtime.sendMessage({ action: 'stopFocus' });
  });
}

function removeBlockingPage(): void {
  const blocker = document.getElementById('mathalama-focus-blocker');
  if (blocker) {
    blocker.remove();
  }
}

function updateTimer(session: FocusSession): void {
  const timerElement = document.getElementById('timer');
  const remainingElement = document.getElementById('remaining');

  if (!timerElement) return;

  const elapsed = Date.now() - session.startTime;
  const remaining = Math.max(0, session.duration - elapsed);
  const minutes = Math.floor(remaining / 60000);
  const seconds = Math.floor((remaining % 60000) / 1000);

  timerElement.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;

  if (remaining > 0) {
    setTimeout(() => updateTimer(session), 1000);
  }
}

// Check on page load
checkAndBlock();

// Check periodically for dynamic content and URL changes
const observer = new MutationObserver(() => {
  checkAndBlock();
});

observer.observe(document.documentElement, {
  childList: true,
  subtree: true
});

// Also listen for dynamic navigation (SPA)
window.addEventListener('hashchange', checkAndBlock);
window.addEventListener('popstate', checkAndBlock);

