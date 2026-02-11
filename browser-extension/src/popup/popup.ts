// Popup script - Shows focus status and allows managing blocked sites

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

document.addEventListener('DOMContentLoaded', async () => {
  await loadSettings();
  await updateUI();
  setupEventListeners();
  
  // Update UI every second while active
  setInterval(updateUI, 1000);
});

async function loadSettings(): Promise<void> {
  const data = await chrome.storage.local.get('defaultBlockedSites');
  const sites = data.defaultBlockedSites || [];

  const siteList = document.getElementById('site-list');
  if (siteList) {
    siteList.innerHTML = sites
      .map((site: BlockedSite) => `
        <label class="site-item">
          <input type="checkbox" data-domain="${site.domain}" ${site.enabled ? 'checked' : ''}>
          <span>${site.domain}</span>
        </label>
      `)
      .join('');

    // Add event listeners to checkboxes
    siteList.querySelectorAll('input[type="checkbox"]').forEach((checkbox: HTMLInputElement) => {
      checkbox.addEventListener('change', saveSitePreferences);
    });
  }
}

function setupEventListeners(): void {
  // Setup any additional listeners if needed
}

async function updateUI(): Promise<void> {
  const data = await chrome.storage.local.get('focusSession');
  const session = data.focusSession as FocusSession;

  const statusBadge = document.getElementById('status-badge');
  const timer = document.getElementById('timer');

  if (session?.active) {
    if (statusBadge) statusBadge.textContent = '● Focus Mode Active';
    if (statusBadge) statusBadge.className = 'status-badge active';
    if (timer) timer.style.display = 'block';

    updateTimer(session);
  } else {
    if (statusBadge) statusBadge.textContent = '○ Not Active';
    if (statusBadge) statusBadge.className = 'status-badge inactive';
    if (timer) timer.style.display = 'none';
  }
}

function updateTimer(session: FocusSession): void {
  const timerEl = document.getElementById('timer');
  if (!timerEl) return;

  const elapsed = Date.now() - session.startTime;
  const remaining = Math.max(0, session.duration - elapsed);
  const minutes = Math.floor(remaining / 60000);
  const seconds = Math.floor((remaining % 60000) / 1000);

  timerEl.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

async function saveSitePreferences(): Promise<void> {
  const data = await chrome.storage.local.get('defaultBlockedSites');
  const sites = data.defaultBlockedSites || [];

  // Update enabled status based on checkboxes
  document.querySelectorAll('input[type="checkbox"]').forEach((checkbox: HTMLInputElement) => {
    const domain = checkbox.getAttribute('data-domain');
    const site = sites.find((s: BlockedSite) => s.domain === domain);
    if (site) {
      site.enabled = checkbox.checked;
    }
  });

  await chrome.storage.local.set({ defaultBlockedSites: sites });
}
