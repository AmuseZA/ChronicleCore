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
const TAB_CHECK_INTERVAL = 1000; // Check active tab every second
const DEBOUNCE_MS = 2000;  // Minimum time between events for same URL
const SERVER_CHECK_INTERVAL = 30000;  // Check server status every 30 seconds

// ============================================
// STATE
// ============================================

let lastEvent = { url: null, time: 0 };
let isTracking = false;
let isPaused = false;
let isStopped = true;
let isServerReachable = false;
let lastServerStatus = null;
let pageStartTime = Date.now();
let currentTabId = null;

// ============================================
// INITIALIZATION
// ============================================

// Check server on startup
checkServerStatus();

// Periodic server check
setInterval(checkServerStatus, SERVER_CHECK_INTERVAL);

// Periodic active tab check (Fallback for missed events, vital for Sidebar/Panels)
setInterval(checkActiveTab, TAB_CHECK_INTERVAL);

// ============================================
// EVENT LISTENERS
// ============================================

// Track tab activation (switching between tabs)
chrome.tabs.onActivated.addListener(async (activeInfo) => {
  if (!isTracking || isPaused || isStopped) return;
  checkActiveTab(true); // Force check
});

// Track page loads within active tab
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (!isTracking || isPaused || isStopped) return;

  // Only track when page finishes loading and it's the active tab
  if (changeInfo.status === 'complete' && tab.active) {
    checkActiveTab(true);
  }
});

// Track window focus changes
chrome.windows.onFocusChanged.addListener(async (windowId) => {
  if (!isTracking || isPaused || isStopped || windowId === chrome.windows.WINDOW_ID_NONE) return;
  checkActiveTab(true);
});

// ============================================
// HELPERS
// ============================================

async function getActiveTab() {
  try {
    // Strategy 1: Last focused window (Native Chrome logic)
    // This is often the most accurate for "what is the user looking at"
    let [tab] = await chrome.tabs.query({ active: true, lastFocusedWindow: true });

    // Strategy 2: Scan ALL windows for one that is focused
    // Opera sidebars/panels might have strange types, so we don't filter by type anymore
    if (!tab) {
      // populating tabs is expensive but necessary here
      const windows = await chrome.windows.getAll({ populate: true });
      for (const win of windows) {
        if (win.focused && win.tabs && win.tabs.length > 0) {
          // Find the active tab in this focused window
          const activeTab = win.tabs.find(t => t.active);
          if (activeTab) {
            tab = activeTab;
            // console.log('ChronicleCore Debug: Found tab in focused window', win.type, activeTab.url);
            break;
          }
          // Fallback: take first tab if none marked active (common in simple popups)
          if (!tab) {
            tab = win.tabs[0];
            // console.log('ChronicleCore Debug: Fallback to first tab', win.type, tab.url);
            break;
          }
        }
      }
    }

    // Strategy 3: (Removed, merged into Strategy 2)

    return tab;
  } catch (e) {
    console.error("Error finding active tab:", e);
    return null;
  }
}

async function checkActiveTab(force = false) {
  if (!isTracking || isPaused || isStopped) return;

  const tab = await getActiveTab();

  if (tab && tab.url) {
    // Debug: Log if we see a sidebar-like URL but filter it
    if (isInternalPage(tab.url)) {
      // console.log('ChronicleCore Debug: Ignoring internal page', tab.url);
    }
  }

  if (!tab || !tab.url) return;

  // If URL changed OR forced check (event fired), send event
  // We use a looser debounce here for the poller
  if (tab.url !== lastEvent.url || force) {
    // If we naturally drifted to a new URL without an event firing
    const now = Date.now();

    // Basic Debounce: Don't spam the same URL if checked recently via polling
    if (!force && tab.url === lastEvent.url && (now - lastEvent.time) < 2000) return;

    // Calculate duration from previous page start
    const duration = now - pageStartTime;
    pageStartTime = now;
    currentTabId = tab.id;

    // Determine event type
    let type = 'PAGE_VIEW';
    if (force) type = 'TAB_ACTIVATED'; // Heuristic

    await sendEvent(tab, type, duration);
  }
}


// Handle messages from popup
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === 'resumeTracking') {
    controlTracking('resume').then(newState => sendResponse(newState));
    return true;
  } else if (message.action === 'startTracking') {
    controlTracking('start').then(newState => sendResponse(newState));
    return true;
  } else if (message.action === 'pauseTracking') {
    controlTracking('pause').then(newState => sendResponse(newState));
    return true;
  } else if (message.action === 'stopTracking') {
    controlTracking('stop').then(newState => sendResponse(newState));
    return true;
  } else if (message.action === 'getStatus') {
    sendResponse({
      isTracking,
      isPaused,
      isStopped,
      isServerReachable,
      lastEvent,
      serverStatus: lastServerStatus
    });
  } else if (message.action === 'checkServer') {
    checkServerStatus().then(() => {
      sendResponse({
        isServerReachable,
        isTracking,
        isPaused,
        isStopped,
        serverStatus: lastServerStatus
      });
    });
    return true; // Indicates async response
  } else if (message.action === 'getDebugInfo') {
    // Explicitly ask for ALL types
    const types = ['normal', 'popup', 'panel', 'app', 'devtools'];
    chrome.windows.getAll({ populate: true, windowTypes: types }, (windows) => {
      // Simplify output for readability
      const debugData = windows.map(w => ({
        id: w.id,
        type: w.type,
        focused: w.focused,
        state: w.state,
        tabs: w.tabs.map(t => ({
          id: t.id,
          active: t.active,
          url: t.url ? (t.url.length > 50 ? t.url.substring(0, 50) + '...' : t.url) : 'no-url',
          title: t.title ? (t.title.length > 30 ? t.title.substring(0, 30) + '...' : t.title) : 'no-title'
        }))
      }));
      sendResponse({
        windows: debugData,
        serverStatus: lastServerStatus,
        lastExtensionEvent: lastEvent,
        timestamp: new Date().toISOString()
      });
    });
    return true;
  } else if (message.action === 'contentActive') {
    // Content script told us it's active!
    // This is the "God Mode" for tracking sidebars that are otherwise invisible

    if (!isTracking || isPaused || isStopped) return;

    const now = Date.now();

    // If the URL is different or it's been a while, log it
    if (message.url !== lastEvent.url || (now - lastEvent.time) > DEBOUNCE_MS) {

      let urlObj;
      try { urlObj = new URL(message.url); } catch (e) { }

      if (urlObj && !isInternalPage(message.url)) {

        // Calculate duration from previous
        const duration = now - pageStartTime;
        pageStartTime = now;

        // We don't have a tab ID if it's a sidebar (sender.tab might be null or partial)
        // But we can still log the event!

        const description = generateDescription(urlObj, message.title);

        const payload = {
          url: message.url,
          title: message.title || '',
          domain: urlObj.hostname,
          description: description,
          tab_id: sender.tab ? sender.tab.id : -1, // -1 for unknown/sidebar
          timestamp: new Date().toISOString(),
          event_type: 'INTERACTION', // New event type implies explicit user action
          duration_ms: duration
        };

        // Send to server
        fetch(`${API_BASE}/api/v1/events/ingest`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        }).then(() => {
          lastEvent = { url: message.url, time: now, description };
          console.log('ChronicleCore: Sidebar/Content Interaction -', description);
        }).catch(err => console.warn('Sync failed', err));
      }
    }
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
 * Check if the ChronicleCore server is reachable AND sync state
 */
async function checkServerStatus() {
  try {
    // First check health
    const healthResp = await fetch(`${API_BASE}/health`, {
      method: 'GET',
      signal: (AbortSignal && AbortSignal.timeout) ? AbortSignal.timeout(5000) : null
    });

    if (!healthResp.ok) throw new Error('Health check failed');

    // Then check tracking status to sync state
    const statusResp = await fetch(`${API_BASE}/api/v1/tracking/status`, {
      method: 'GET',
      signal: (AbortSignal && AbortSignal.timeout) ? AbortSignal.timeout(5000) : null
    });

    if (statusResp.ok) {
      const status = await statusResp.json();
      // Sync local state with server state
      isServerReachable = true;

      const state = status.state.toLowerCase(); // Server returns "STOPPED", "ACTIVE", etc.
      isTracking = (state === 'active');
      isPaused = (state === 'paused');
      isStopped = (state === 'stopped');

      // Store full status for popup (Timer, Idle, System Activity)
      lastServerStatus = status;

      console.log(`ChronicleCore: Synced state - ${status.state}`);
    } else {
      isServerReachable = false;
      lastServerStatus = null;
    }

  } catch (err) {
    isServerReachable = false;
    lastServerStatus = null;
    console.log('ChronicleCore: Server not reachable');
  }

  return isServerReachable;
}

/**
 * Control tracking state via API
 */
async function controlTracking(action) {
  if (!isServerReachable) return { isTracking, isPaused, isStopped, isServerReachable };

  try {
    const response = await fetch(`${API_BASE}/api/v1/tracking/${action}`, {
      method: 'POST'
    });

    if (response.ok) {
      // Immediate state update (will be confirmed by next poll)
      if (action === 'resume' || action === 'start') {
        isTracking = true; isPaused = false; isStopped = false;
      } else if (action === 'pause') {
        isTracking = false; isPaused = true; isStopped = false;
      } else if (action === 'stop') {
        isTracking = false; isPaused = false; isStopped = true;
      }

      // Force a sync to be sure
      checkServerStatus();
    }
  } catch (err) {
    console.error(`Failed to ${action} tracking:`, err);
  }

  return { isTracking, isPaused, isStopped, isServerReachable };
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
