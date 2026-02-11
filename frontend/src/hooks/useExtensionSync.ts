/**
 * Hook to sync focus session state with browser extension
 * Uses localStorage as a bridge to communicate with content script
 */

export function useExtensionSync() {
  const notifyExtension = (state: 'active' | 'paused' | 'completed') => {
    // Use localStorage to communicate with content script
    localStorage.setItem('focus-session-state', JSON.stringify({
      state,
      timestamp: Date.now()
    }));
    
    // Trigger storage event (for same-tab communication)
    window.dispatchEvent(new Event('focus-session-changed'));
  };

  return { notifyExtension };
}
