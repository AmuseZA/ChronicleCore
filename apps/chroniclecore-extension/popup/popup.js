/**
 * ChronicleCore Extension Popup Script
 */

// DOM Elements
const serverStatus = document.getElementById('serverStatus');
const trackingStatus = document.getElementById('trackingStatus');
const toggleBtn = document.getElementById('toggleBtn');
const toggleIcon = document.getElementById('toggleIcon');
const toggleText = document.getElementById('toggleText');
const dashboardBtn = document.getElementById('dashboardBtn');
const lastActivityText = document.getElementById('lastActivityText');

// State
let isTracking = true;
let isServerReachable = false;

// ============================================
// INITIALIZATION
// ============================================

document.addEventListener('DOMContentLoaded', () => {
  // Get current status from background script
  chrome.runtime.sendMessage({ action: 'getStatus' }, (response) => {
    if (response) {
      isTracking = response.isTracking;
      isServerReachable = response.isServerReachable;

      if (response.lastEvent && response.lastEvent.description) {
        lastActivityText.textContent = response.lastEvent.description;
      }

      updateUI();
    }
  });

  // Check server status
  checkServer();
});

// ============================================
// EVENT HANDLERS
// ============================================

toggleBtn.addEventListener('click', () => {
  chrome.runtime.sendMessage({ action: 'toggleTracking' }, (response) => {
    if (response) {
      isTracking = response.isTracking;
      isServerReachable = response.isServerReachable;
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

function updateUI() {
  // Update server status
  updateServerStatus();

  // Update tracking status
  if (isTracking) {
    trackingStatus.className = 'status status-active';
    trackingStatus.querySelector('.status-text').textContent = 'Tracking active';
    toggleBtn.classList.remove('paused');
    toggleIcon.innerHTML = '&#9208;'; // Pause icon
    toggleText.textContent = 'Pause Tracking';
  } else {
    trackingStatus.className = 'status status-paused';
    trackingStatus.querySelector('.status-text').textContent = 'Tracking paused';
    toggleBtn.classList.add('paused');
    toggleIcon.innerHTML = '&#9654;'; // Play icon
    toggleText.textContent = 'Resume Tracking';
  }

  // Disable toggle if server not reachable
  toggleBtn.disabled = !isServerReachable;
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
      isServerReachable = response.isServerReachable;
      updateUI();
    }
  });
}
