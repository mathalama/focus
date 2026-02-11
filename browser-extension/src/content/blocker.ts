// Logic to block distracting sites during focus mode
interface FocusSession {
  active: boolean;
  startTime: number;
  duration: number;
  blockedSites: Array<{ domain: string; pattern: string; enabled: boolean }>;
}

export async function checkAndBlock(): Promise<void> {
  const data = await chrome.storage.local.get('focusSession');
  const session = data.focusSession as FocusSession;

  if (!session.active) {
    // No active focus session, remove any blocking page
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
  // Check if blocking page already exists
  if (document.getElementById('mathalama-focus-blocker')) {
    return;
  }

  // Create blocking overlay
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
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 999999;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      }
      
      .blocker-content {
        text-align: center;
        color: white;
        padding: 40px;
        border-radius: 20px;
        background: rgba(0, 0, 0, 0.1);
        max-width: 600px;
      }
      
      .blocker-title {
        font-size: 48px;
        font-weight: 700;
        margin-bottom: 20px;
      }
      
      .blocker-message {
        font-size: 24px;
        margin-bottom: 30px;
        opacity: 0.95;
      }
      
      .blocker-timer {
        font-size: 64px;
        font-weight: 300;
        margin: 30px 0;
        font-variant-numeric: tabular-nums;
      }
      
      .blocker-remaining {
        font-size: 18px;
        margin-bottom: 40px;
        opacity: 0.8;
      }
      
      .blocker-button {
        background: rgba(255, 255, 255, 0.25);
        color: white;
        border: 2px solid white;
        padding: 12px 32px;
        border-radius: 8px;
        font-size: 16px;
        cursor: pointer;
        transition: all 0.3s ease;
        margin: 0 10px;
      }
      
      .blocker-button:hover {
        background: rgba(255, 255, 255, 0.4);
        transform: scale(1.05);
      }
      
      .blocker-button.danger {
        background: rgba(255, 107, 107, 0.3);
        border-color: #FF6B6B;
      }
      
      .blocker-button.danger:hover {
        background: rgba(255, 107, 107, 0.5);
      }
    </style>
    
    <div class="blocker-content">
      <div class="blocker-title">🎯 Focus Mode Active</div>
      <div class="blocker-message">Stay focused, avoid distractions!</div>
      <div class="blocker-timer" id="timer">25:00</div>
      <div class="blocker-remaining" id="remaining">Time remaining in focus session</div>
      <div>
        <button class="blocker-button" onclick="window.history.back()">Go Back</button>
        <button class="blocker-button danger" id="end-focus">End Focus</button>
      </div>
    </div>
  `;

  document.documentElement.appendChild(blocker);

  // Update timer
  updateTimer(session);

  // Handle end focus button
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
  } else if (remainingElement) {
    remainingElement.textContent = 'Focus session complete!';
  }
}
