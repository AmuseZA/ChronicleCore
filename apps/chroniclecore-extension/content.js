/**
 * ChronicleCore Content Script
 * 
 * Runs in all frames/tabs to detect activity even if the browser
 * hides the tab from the main 'tabs' API (e.g. Opera Sidebars).
 */

// Debounce checks
let lastActivity = 0;
const DEBOUNCE = 1000;

function notifyBackground() {
    const now = Date.now();
    if (now - lastActivity < DEBOUNCE) return;
    lastActivity = now;

    // Send "I am active" signal
    chrome.runtime.sendMessage({
        action: 'contentActive',
        url: window.location.href,
        title: document.title
    }).catch(err => {
        // Ignore errors (e.g. extension context invalidated during reload)
    });
}

// Listen for interactions
window.addEventListener('focus', notifyBackground);
window.addEventListener('click', notifyBackground);
window.addEventListener('keydown', notifyBackground);
window.addEventListener('scroll', notifyBackground);

// Initial check
if (document.hasFocus()) {
    notifyBackground();
}
