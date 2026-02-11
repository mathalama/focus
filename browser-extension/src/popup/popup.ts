// Popup script for focus mode control
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

let currentDuration = 25;
let updateInterval: number | null = null;

document.addEventListener('DOMContentLoaded', async () => {
  await loadSettings();
  await updateUI();
  setupEventListeners();
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

async function updateUI(): Promise<void> {
  const response = await chrome.runtime.sendMessage({ action: 'getFocusStatus' });
  const { focus, elapsed } = response;

  const statusBadge = document.getElementById('status-badge');
  const timer = document.getElementById('timer');
  const durationInput = document.getElementById('duration-input');
  const startBtn = document.getElementById('start-btn');
  const stopBtn = document.getElementById('stop-btn');

  if (focus.active) {
    if (statusBadge) statusBadge.textContent = '🔴 Active';
    if (statusBadge) statusBadge.className = 'status-badge active';
    if (timer) timer.style.display = 'block';
    if (durationInput) durationInput.style.display = 'none';
    if (startBtn) startBtn.style.display = 'none';
    if (stopBtn) stopBtn.style.display = 'flex';

    updateTimer(focus, elapsed);
  } else {
    if (statusBadge) statusBadge.textContent = 'Not Active';
    if (statusBadge) statusBadge.className = 'status-badge inactive';
    if (timer) timer.style.display = 'none';
    if (durationInput) durationInput.style.display = 'block';
    if (startBtn) startBtn.style.display = 'flex';
    if (stopBtn) stopBtn.style.display = 'none';

    if (updateInterval !== null) {
      clearInterval(updateInterval);
      updateInterval = null;
    }
  }
}

function updateTimer(focus: FocusSession, elapsed: number): void {
  const timerEl = document.getElementById('timer');
  if (!timerEl) return;

  const remaining = Math.max(0, focus.duration - elapsed);
  const minutes = Math.floor(remaining / 60000);
  const seconds = Math.floor((remaining % 60000) / 1000);

  timerEl.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;

  if (remaining <= 0) {
    updateUI(); // Session ended
  } else if (updateInterval === null) {
    updateInterval = setInterval(() => {
      chrome.runtime.sendMessage({ action: 'getFocusStatus' }, (response) => {
        updateTimer(response.focus, response.elapsed);
      });
    }, 1000) as unknown as number;
  }
}

function setupEventListeners(): void {
  // Preset duration buttons
  document.querySelectorAll('.preset-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      const duration = target.getAttribute('data-duration');
      if (duration) {
        currentDuration = parseInt(duration, 10);
        document.querySelectorAll('.preset-btn').forEach((b) => b.classList.remove('active'));
        target.classList.add('active');
      }
    });
  });

  // Start focus button
  document.getElementById('start-btn')?.addEventListener('click', startFocus);

  // Stop focus button
  document.getElementById('stop-btn')?.addEventListener('click', stopFocus);
}

async function startFocus(): Promise<void> {
  const data = await chrome.storage.local.get('defaultBlockedSites');
  const blockedSites = (data.defaultBlockedSites || []).filter(
    (site: BlockedSite) => site.enabled
  );

  await chrome.runtime.sendMessage({
    action: 'startFocus',
    duration: currentDuration,
    blockedSites: blockedSites
  });

  await updateUI();
}

async function stopFocus(): Promise<void> {
  await chrome.runtime.sendMessage({ action: 'stopFocus' });
  await updateUI();
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
  await chrome.runtime.sendMessage({
    action: 'updateBlockedSites',
    sites: sites.filter((s: BlockedSite) => s.enabled)
  });
}

declare global {
  function setCustomDuration(): void;
}

(window as any).setCustomDuration = async function (): Promise<void> {
  const input = document.getElementById('custom-duration') as HTMLInputElement;
  const duration = parseInt(input.value, 10);

  if (duration > 0 && duration <= 480) {
    currentDuration = duration;
  } else {
    alert('Duration must be between 1 and 480 minutes');
  }
};
