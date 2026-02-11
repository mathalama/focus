// Content script that runs on all pages to enforce blocking
import { checkAndBlock } from './blocker.js';

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
