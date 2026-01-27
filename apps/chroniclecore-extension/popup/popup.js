/**
 * ChronicleCore Extension Popup Script
 */

// DOM Elements
const serverStatus = document.getElementById('serverStatus');
const trackingStatus = document.getElementById('trackingStatus');
const resumeBtn = document.getElementById('resumeBtn');
const pauseBtn = document.getElementById('pauseBtn');
const stopBtn = document.getElementById('stopBtn');
const dashboardBtn = document.getElementById('dashboardBtn');
const lastActivityText = document.getElementById('lastActivityText');
const systemActivityText = document.getElementById('systemActivityText');
const systemActivityContainer = document.getElementById('systemActivity');
const dailyTimeText = document.getElementById('dailyTime');
const idleIndicator = document.getElementById('idleIndicator');
const idleText = document.getElementById('idleText');
const themeToggle = document.getElementById('themeToggle');
const versionEl = document.querySelector('.version'); // Select version element
const debugSection = document.getElementById('debugSection');
const debugInfo = document.getElementById('debugInfo');

// State
let isTracking = false;
let isPaused = false;
let isStopped = true;
let isServerReachable = false;
let isDarkMode = false;

// ============================================
// INITIALIZATION
// ============================================

document.addEventListener('DOMContentLoaded', () => {
  // Easter egg: Click version to toggle debug
  if (versionEl) {
    versionEl.style.cursor = 'pointer';
    versionEl.title = 'Click for Debug Info';
    versionEl.addEventListener('click', () => {
      if (debugSection.style.display === 'none') {
        debugSection.style.display = 'block';
        fetchDebugInfo();
      } else {
        debugSection.style.display = 'none';
      }
    });
  }
  // Load dark mode preference
  chrome.storage.local.get(['theme'], (result) => {
    if (result.theme === 'dark') {
      enableDarkMode();
    }
  });

  // Get status
  chrome.runtime.sendMessage({ action: 'getStatus' }, (response) => {
    if (response) {
      updateState(response);
      updateUI();
    }
  });

  // Check server status
  checkServer();
});

// ============================================
// EVENT HANDLERS
// ============================================

themeToggle.addEventListener('click', () => {
  isDarkMode = document.body.classList.toggle('dark-mode');
  chrome.storage.local.set({ theme: isDarkMode ? 'dark' : 'light' });
});

resumeBtn.addEventListener('click', () => {
  const action = isStopped ? 'startTracking' : 'resumeTracking';
  chrome.runtime.sendMessage({ action: action }, (response) => {
    if (response) {
      updateState(response);
      updateUI();
    }
  });
});

pauseBtn.addEventListener('click', () => {
  chrome.runtime.sendMessage({ action: 'pauseTracking' }, (response) => {
    if (response) {
      updateState(response);
      updateUI();
    }
  });
});

stopBtn.addEventListener('click', () => {
  chrome.runtime.sendMessage({ action: 'stopTracking' }, (response) => {
    if (response) {
      updateState(response);
      updateUI();
    }
  });
});

dashboardBtn.addEventListener('click', () => {
  chrome.tabs.create({ url: 'http://127.0.0.1:8080' });
  window.close();
});

// ============================================
// UI UPDATES
// ============================================

function updateState(response) {
  isTracking = response.isTracking;
  isPaused = response.isPaused;
  isStopped = response.isStopped;
  isServerReachable = response.isServerReachable;

  // Last browser activity
  if (response.lastEvent && response.lastEvent.description) {
    lastActivityText.textContent = response.lastEvent.description;
  }

  // System Activity & Daily Time (from server sync in background.js)
  // We need to fetch this from background, which gets it from server.
  // The background script's 'getStatus' needs to effectively proxy or cache the server status.
  // Currently background sends: { isTracking, isPaused, isStopped, isServerReachable, lastEvent }
  // We need to add: { currentWindow, idleSeconds, todayTimeSeconds }

  // Actually, background.js 'checkServerStatus' logs it but doesn't persist all details to a global var we can read.
  // Let's rely on the background script to forward the latest server status if we ask for it.
  // However, for now, response might NOT have these new fields until we update background.js to store them.
  // Assuming background.js is updated (next step), we use them:

  if (response.serverStatus) {
    updateExtendedStatus(response.serverStatus);
  }
}

function updateExtendedStatus(status) {
  // System Activity
  if (status.current_window) {
    systemActivityContainer.style.display = 'block';
    systemActivityText.textContent = `${status.current_window.app_name}: ${status.current_window.title}`;
  } else {
    systemActivityContainer.style.display = 'none';
  }

  // Daily Time
  if (status.today_time_seconds !== undefined) {
    dailyTimeText.textContent = formatDuration(status.today_time_seconds);
  }

  // Idle Indicator
  if (status.idle_seconds > 60) { // Show if idle > 1 min
    idleIndicator.style.display = 'flex';
    idleText.textContent = `Idle for ${formatDuration(status.idle_seconds)}`;
  } else {
    idleIndicator.style.display = 'none';
  }
}

function updateUI() {
  updateServerStatus();

  // Reset display
  resumeBtn.style.display = 'none';
  pauseBtn.style.display = 'none';
  stopBtn.style.display = 'none';

  if (!isServerReachable) {
    trackingStatus.className = 'status status-error';
    trackingStatus.querySelector('.status-text').textContent = 'Cannot control tracking';
  } else if (isStopped) {
    trackingStatus.className = 'status status-paused';
    trackingStatus.querySelector('.status-text').textContent = 'Tracking stopped';
    resumeBtn.style.display = 'flex';
  } else if (isPaused) {
    trackingStatus.className = 'status status-paused';
    trackingStatus.querySelector('.status-text').textContent = 'Tracking paused';
    resumeBtn.style.display = 'flex';
    stopBtn.style.display = 'flex';
  } else {
    trackingStatus.className = 'status status-active';
    trackingStatus.querySelector('.status-text').textContent = 'Tracking active';
    pauseBtn.style.display = 'flex';
    stopBtn.style.display = 'flex';
  }
}

function updateServerStatus() {
  if (isServerReachable) {
    serverStatus.className = 'status status-active';
    serverStatus.querySelector('.status-text').textContent = 'Server connected';
  } else {
    serverStatus.className = 'status status-error';
    serverStatus.querySelector('.status-text').textContent = 'Server not reachable';
  }
}

function checkServer() {
  serverStatus.className = 'status status-checking';
  serverStatus.querySelector('.status-text').textContent = 'Checking server...';

  chrome.runtime.sendMessage({ action: 'checkServer' }, (response) => {
    if (response) {
      updateState(response);
      updateUI();
    }
  });
}

function enableDarkMode() {
  document.body.classList.add('dark-mode');
  isDarkMode = true;
}

function formatDuration(seconds) {
  if (!seconds) return '0m';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

function fetchDebugInfo() {
  debugInfo.textContent = "Loading windows...";
  chrome.runtime.sendMessage({ action: 'getDebugInfo' }, (response) => {
    debugInfo.textContent = JSON.stringify(response, null, 2);
  });
}
