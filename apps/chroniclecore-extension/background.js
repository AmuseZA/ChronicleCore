/**
 * ChronicleCore Browser Extension - Background Service Worker
 *
 * Tracks browser activity and sends events to the local ChronicleCore server.
 * Generates human-readable descriptions for activities.
 */

// Import patterns (for MV3 module support)
importScripts('patterns.js');

// ============================================
// CONFIGURATION
// ============================================

const API_BASE = 'http://127.0.0.1:8080';
const DEBOUNCE_MS = 2000;  // Minimum time between events for same URL
const SERVER_CHECK_INTERVAL = 30000;  // Check server status every 30 seconds

// ============================================
// STATE
// ============================================

let lastEvent = { url: null, time: 0 };
let isTracking = true;
let isServerReachable = false;
let pageStartTime = Date.now();
let currentTabId = null;

// ============================================
// INITIALIZATION
// ============================================

// Load saved state
chrome.storage.local.get(['isTracking'], (result) => {
  if (result.isTracking !== undefined) {
    isTracking = result.isTracking;
  }
});

// Check server on startup
checkServerStatus();

// Periodic server check
setInterval(checkServerStatus, SERVER_CHECK_INTERVAL);

// ============================================
// EVENT LISTENERS
// ============================================

// Track tab activation (switching between tabs)
chrome.tabs.onActivated.addListener(async (activeInfo) => {
  if (!isTracking) return;

  try {
    const tab = await chrome.tabs.get(activeInfo.tabId);
    const duration = Date.now() - pageStartTime;
    pageStartTime = Date.now();
    currentTabId = activeInfo.tabId;

    await sendEvent(tab, 'TAB_ACTIVATED', duration);
  } catch (err) {
    console.warn('ChronicleCore: Tab activation error', err.message);
  }
});

// Track page loads within active tab
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (!isTracking) return;

  // Only track when page finishes loading and it's the active tab
  if (changeInfo.status === 'complete' && tab.active) {
    pageStartTime = Date.now();
    currentTabId = tabId;
    sendEvent(tab, 'PAGE_LOADED', null);
  }
});

// Track window focus changes
chrome.windows.onFocusChanged.addListener(async (windowId) => {
  if (!isTracking || windowId === chrome.windows.WINDOW_ID_NONE) return;

  try {
    const [tab] = await chrome.tabs.query({ active: true, windowId: windowId });
    if (tab) {
      pageStartTime = Date.now();
      currentTabId = tab.id;
      sendEvent(tab, 'WINDOW_FOCUSED', null);
    }
  } catch (err) {
    console.warn('ChronicleCore: Window focus error', err.message);
  }
});

// Handle messages from popup
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === 'toggleTracking') {
    isTracking = !isTracking;
    chrome.storage.local.set({ isTracking: isTracking });
    sendResponse({ isTracking, isServerReachable });
  } else if (message.action === 'getStatus') {
    sendResponse({ isTracking, isServerReachable, lastEvent });
  } else if (message.action === 'checkServer') {
    checkServerStatus().then(() => {
      sendResponse({ isServerReachable });
    });
    return true; // Indicates async response
  }
});

// ============================================
// CORE FUNCTIONS
// ============================================

/**
 * Send an event to the ChronicleCore server
 */
async function sendEvent(tab, eventType, durationMs) {
  // Skip invalid tabs
  if (!tab || !tab.url) return;

  // Skip internal browser pages
  if (isInternalPage(tab.url)) return;

  // Debounce rapid events to same URL
  const now = Date.now();
  if (tab.url === lastEvent.url && (now - lastEvent.time) < DEBOUNCE_MS) {
    return;
  }

  // Skip if server is not reachable
  if (!isServerReachable) {
    console.log('ChronicleCore: Server not reachable, skipping event');
    return;
  }

  let url;
  try {
    url = new URL(tab.url);
  } catch (e) {
    console.warn('ChronicleCore: Invalid URL', tab.url);
    return;
  }

  // Generate human-readable description
  const description = generateDescription(url, tab.title);

  const payload = {
    url: tab.url,
    title: tab.title || '',
    domain: url.hostname,
    description: description,
    tab_id: tab.id,
    timestamp: new Date().toISOString(),
    event_type: eventType,
    duration_ms: durationMs
  };

  try {
    const response = await fetch(`${API_BASE}/api/v1/events/ingest`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (response.ok) {
      lastEvent = { url: tab.url, time: now, description };
      console.log('ChronicleCore:', eventType, '-', description);
    } else {
      console.warn('ChronicleCore: Server returned', response.status);
    }
  } catch (err) {
    console.warn('ChronicleCore: Failed to send event', err.message);
    isServerReachable = false;
  }
}

/**
 * Check if the ChronicleCore server is reachable
 */
async function checkServerStatus() {
  try {
    const response = await fetch(`${API_BASE}/health`, {
      method: 'GET',
      signal: AbortSignal.timeout(5000)
    });

    isServerReachable = response.ok;

    if (isServerReachable) {
      console.log('ChronicleCore: Server is reachable');
    }
  } catch (err) {
    isServerReachable = false;
    console.log('ChronicleCore: Server not reachable');
  }

  return isServerReachable;
}

/**
 * Check if a URL is an internal browser page
 */
function isInternalPage(url) {
  if (!url) return true;

  const internalPrefixes = [
    'chrome://',
    'chrome-extension://',
    'edge://',
    'about:',
    'moz-extension://',
    'brave://',
    'opera://',
    'vivaldi://'
  ];

  return internalPrefixes.some(prefix => url.startsWith(prefix));
}

// ============================================
// STARTUP
// ============================================

console.log('ChronicleCore Extension loaded');
