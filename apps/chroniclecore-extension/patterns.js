/**
 * Activity Pattern Definitions
 *
 * These patterns extract human-readable descriptions from browser URLs and titles.
 * Only the description is stored - full URLs and message content are not retained.
 */

const PATTERNS = [
  // ============================================
  // MESSAGING APPS
  // ============================================

  // WhatsApp - extract contact from title "John Smith - WhatsApp"
  {
    match: (url) => url.hostname === 'web.whatsapp.com',
    describe: (url, title) => {
      if (!title) return 'On WhatsApp';
      const contact = title.split(' - ')[0];
      if (contact && contact !== 'WhatsApp') {
        return `Chatted with ${contact} on WhatsApp`;
      }
      return 'On WhatsApp';
    }
  },

  // Slack - extract channel/DM from title "#general | Slack"
  {
    match: (url) => url.hostname.includes('slack.com'),
    describe: (url, title) => {
      if (!title) return 'On Slack';
      const channel = title.split(' | ')[0];
      if (channel && channel !== 'Slack') {
        return `Chatted in ${channel} on Slack`;
      }
      return 'On Slack';
    }
  },

  // Microsoft Teams
  {
    match: (url) => url.hostname.includes('teams.microsoft.com'),
    describe: (url, title) => {
      if (!title) return 'On Microsoft Teams';
      const context = title.split(' | ')[0];
      if (context && !context.includes('Microsoft Teams')) {
        return `On Teams: ${context}`;
      }
      return 'On Microsoft Teams';
    }
  },

  // Discord
  {
    match: (url) => url.hostname.includes('discord.com'),
    describe: (url, title) => {
      if (!title) return 'On Discord';
      // Title format: "Server | #channel | Discord" or "@user | Discord"
      const parts = title.split(' | ');
      if (parts.length >= 2 && parts[parts.length - 1].includes('Discord')) {
        if (parts.length === 3) {
          return `Chatted in ${parts[1]} on Discord`;
        }
        return `Chatted on Discord: ${parts[0]}`;
      }
      return 'On Discord';
    }
  },

  // Telegram
  {
    match: (url) => url.hostname.includes('web.telegram.org'),
    describe: (url, title) => {
      if (!title) return 'On Telegram';
      const contact = title.replace(' – Telegram', '').replace(' - Telegram', '');
      if (contact && contact !== 'Telegram') {
        return `Chatted with ${contact} on Telegram`;
      }
      return 'On Telegram';
    }
  },

  // ============================================
  // EMAIL
  // ============================================

  // Gmail - extract subject from title "Subject - email@example.com - Gmail"
  {
    match: (url) => url.hostname === 'mail.google.com',
    describe: (url, title) => {
      if (!title) return 'Checked Gmail';
      if (title.includes(' - ') && !title.startsWith('Inbox')) {
        const subject = title.split(' - ')[0];
        if (subject && subject !== 'Gmail') {
          return `Email: ${subject.substring(0, 60)}`;
        }
      }
      if (title.startsWith('Inbox')) return 'Checked Gmail inbox';
      if (title.includes('Compose')) return 'Composing email in Gmail';
      return 'Checked Gmail';
    }
  },

  // Outlook
  {
    match: (url) => url.hostname.includes('outlook.office.com') || url.hostname.includes('outlook.live.com'),
    describe: (url, title) => {
      if (!title) return 'Checked Outlook';
      if (title.includes(' - Outlook')) {
        const subject = title.replace(' - Outlook', '');
        if (subject && subject !== 'Mail') {
          return `Email: ${subject.substring(0, 60)}`;
        }
      }
      return 'Checked Outlook';
    }
  },

  // ============================================
  // DOCUMENTS & PRODUCTIVITY
  // ============================================

  // Google Docs
  {
    match: (url) => url.hostname === 'docs.google.com' && url.pathname.includes('/document/'),
    describe: (url, title) => {
      if (!title) return 'Edited Google Doc';
      const docName = title.replace(' - Google Docs', '');
      if (docName) {
        return `Edited document: ${docName.substring(0, 50)}`;
      }
      return 'Edited Google Doc';
    }
  },

  // Google Sheets
  {
    match: (url) => url.hostname === 'docs.google.com' && url.pathname.includes('/spreadsheets/'),
    describe: (url, title) => {
      if (!title) return 'Edited Google Sheet';
      const sheetName = title.replace(' - Google Sheets', '');
      if (sheetName) {
        return `Edited spreadsheet: ${sheetName.substring(0, 50)}`;
      }
      return 'Edited Google Sheet';
    }
  },

  // Google Slides
  {
    match: (url) => url.hostname === 'docs.google.com' && url.pathname.includes('/presentation/'),
    describe: (url, title) => {
      if (!title) return 'Edited Google Slides';
      const slideName = title.replace(' - Google Slides', '');
      if (slideName) {
        return `Edited presentation: ${slideName.substring(0, 50)}`;
      }
      return 'Edited Google Slides';
    }
  },

  // Notion
  {
    match: (url) => url.hostname.includes('notion.so'),
    describe: (url, title) => {
      if (!title) return 'Used Notion';
      const page = title.replace(' | Notion', '').replace(' - Notion', '');
      if (page && page !== 'Notion') {
        return `Edited: ${page.substring(0, 50)} in Notion`;
      }
      return 'Used Notion';
    }
  },

  // Confluence
  {
    match: (url) => url.hostname.includes('atlassian.net') && url.pathname.includes('/wiki/'),
    describe: (url, title) => {
      if (!title) return 'Used Confluence';
      return `Viewed Confluence: ${title.substring(0, 50)}`;
    }
  },

  // ============================================
  // CODE & DEVELOPMENT
  // ============================================

  // GitHub Pull Requests
  {
    match: (url) => url.hostname === 'github.com' && url.pathname.includes('/pull/'),
    describe: (url, title) => {
      const match = url.pathname.match(/\/([^\/]+)\/([^\/]+)\/pull\/(\d+)/);
      if (match) {
        return `Reviewed PR #${match[3]} on ${match[1]}/${match[2]}`;
      }
      return 'Reviewed pull request on GitHub';
    }
  },

  // GitHub Issues
  {
    match: (url) => url.hostname === 'github.com' && url.pathname.includes('/issues/'),
    describe: (url, title) => {
      const match = url.pathname.match(/\/([^\/]+)\/([^\/]+)\/issues\/(\d+)/);
      if (match) {
        return `Viewed issue #${match[3]} on ${match[1]}/${match[2]}`;
      }
      return 'Viewed issue on GitHub';
    }
  },

  // GitHub Repository
  {
    match: (url) => url.hostname === 'github.com' && !url.pathname.includes('/pull/') && !url.pathname.includes('/issues/'),
    describe: (url, title) => {
      const match = url.pathname.match(/^\/([^\/]+)\/([^\/]+)/);
      if (match && match[1] && match[2]) {
        if (url.pathname.includes('/actions')) return `Viewed CI/CD for ${match[1]}/${match[2]}`;
        if (url.pathname.includes('/commits')) return `Viewed commits for ${match[1]}/${match[2]}`;
        return `Browsed ${match[1]}/${match[2]} on GitHub`;
      }
      return 'Browsed GitHub';
    }
  },

  // GitLab
  {
    match: (url) => url.hostname.includes('gitlab.com') || url.hostname.includes('gitlab'),
    describe: (url, title) => {
      if (url.pathname.includes('/merge_requests/')) {
        return 'Reviewed merge request on GitLab';
      }
      if (url.pathname.includes('/issues/')) {
        return 'Viewed issue on GitLab';
      }
      return 'Browsed GitLab';
    }
  },

  // Jira
  {
    match: (url) => url.hostname.includes('atlassian.net') && url.pathname.includes('/browse/'),
    describe: (url, title) => {
      const match = url.pathname.match(/\/browse\/([A-Z]+-\d+)/);
      if (match) {
        return `Viewed Jira ticket ${match[1]}`;
      }
      return 'Viewed Jira ticket';
    }
  },

  // ============================================
  // DESIGN
  // ============================================

  // Figma
  {
    match: (url) => url.hostname.includes('figma.com'),
    describe: (url, title) => {
      if (!title) return 'Used Figma';
      // Title format: "Design Name – Figma"
      const project = title.split(' – ')[0] || title.split(' - ')[0];
      if (project && project !== 'Figma') {
        return `Designed: ${project.substring(0, 50)} in Figma`;
      }
      return 'Used Figma';
    }
  },

  // Canva
  {
    match: (url) => url.hostname.includes('canva.com'),
    describe: (url, title) => {
      if (!title) return 'Used Canva';
      return `Designed in Canva`;
    }
  },

  // ============================================
  // VIDEO & MEDIA
  // ============================================

  // YouTube - watching videos
  {
    match: (url) => url.hostname.includes('youtube.com') && url.pathname === '/watch',
    describe: (url, title) => {
      if (!title) return 'Watched YouTube';
      const video = title.replace(' - YouTube', '');
      if (video) {
        return `Watched: ${video.substring(0, 60)}`;
      }
      return 'Watched YouTube';
    }
  },

  // YouTube - browsing
  {
    match: (url) => url.hostname.includes('youtube.com') && url.pathname !== '/watch',
    describe: (url, title) => {
      return 'Browsed YouTube';
    }
  },

  // ============================================
  // SOCIAL & PROFESSIONAL
  // ============================================

  // LinkedIn
  {
    match: (url) => url.hostname.includes('linkedin.com'),
    describe: (url, title) => {
      if (url.pathname.includes('/messaging')) {
        return 'Messaging on LinkedIn';
      }
      if (url.pathname.includes('/jobs')) {
        return 'Viewing jobs on LinkedIn';
      }
      if (url.pathname.includes('/in/')) {
        return 'Viewed LinkedIn profile';
      }
      return 'Browsed LinkedIn';
    }
  },

  // Twitter/X
  {
    match: (url) => url.hostname.includes('twitter.com') || url.hostname.includes('x.com'),
    describe: (url, title) => {
      if (url.pathname.includes('/messages')) {
        return 'Messaging on X';
      }
      return 'Browsed X (Twitter)';
    }
  },

  // ============================================
  // CHRONICLECORE (SELF)
  // ============================================

  {
    match: (url) => url.hostname === '127.0.0.1' || url.hostname === 'localhost',
    describe: (url, title) => {
      if (url.pathname.includes('/profiles')) return 'Managed profiles in ChronicleCore';
      if (url.pathname.includes('/review')) return 'Reviewed time entries in ChronicleCore';
      if (url.pathname.includes('/rules')) return 'Configured rules in ChronicleCore';
      if (url.pathname.includes('/settings')) return 'Changed settings in ChronicleCore';
      if (url.pathname.includes('/suggestions')) return 'Reviewed suggestions in ChronicleCore';
      return 'Used ChronicleCore';
    }
  }
];

/**
 * Generate a human-readable description from URL and title
 * @param {URL} url - Parsed URL object
 * @param {string} title - Page title
 * @returns {string} Human-readable activity description
 */
function generateDescription(url, title) {
  // Try each pattern
  for (const pattern of PATTERNS) {
    try {
      if (pattern.match(url)) {
        return pattern.describe(url, title);
      }
    } catch (e) {
      console.warn('Pattern match error:', e);
    }
  }

  // Default fallback: "Browsed example.com: Page Title"
  const shortTitle = title ? title.substring(0, 50) : '';
  const domain = url.hostname.replace('www.', '');
  return `Browsed ${domain}${shortTitle ? ': ' + shortTitle : ''}`;
}

// Export for ES modules (Chrome MV3)
if (typeof globalThis !== 'undefined') {
  globalThis.generateDescription = generateDescription;
}

// Export for Firefox MV2
if (typeof window !== 'undefined') {
  window.generateDescription = generateDescription;
}
