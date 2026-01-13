import express from 'express';
import cors from 'cors';

const app = express();
const PORT = 8080;

app.use(cors({
  origin: ['http://localhost:5173', 'http://127.0.0.1:5173', 'http://localhost:3000'],
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type']
}));

app.use(express.json());

// Mock data
let trackingState = 'STOPPED';
let lastActiveAt = new Date().toISOString();

const mockBlocks = [
  {
    block_id: 1,
    ts_start: '2026-01-09T09:00:00Z',
    ts_end: '2026-01-09T10:30:00Z',
    duration_minutes: 90.0,
    duration_hours: 1.5,
    primary_app_name: 'EXCEL.EXE',
    primary_domain: null,
    title_summary: 'Budget 2026.xlsx',
    profile_id: 1,
    client_name: 'Acme Corp',
    project_name: null,
    service_name: 'Bookkeeping',
    confidence: 'HIGH',
    billable: true,
    locked: false,
    notes: null,
    description: 'Excel - Budget 2026.xlsx',
    created_at: '2026-01-09T10:30:05Z',
    updated_at: '2026-01-09T10:30:05Z'
  },
  {
    block_id: 2,
    ts_start: '2026-01-09T10:30:00Z',
    ts_end: '2026-01-09T11:15:00Z',
    duration_minutes: 45.0,
    duration_hours: 0.75,
    primary_app_name: 'msedge.exe',
    primary_domain: 'github.com',
    title_summary: 'Pull Requests - GitHub',
    profile_id: null,
    client_name: null,
    project_name: null,
    service_name: null,
    confidence: 'LOW',
    billable: true,
    locked: false,
    notes: null,
    description: 'browsing github.com',
    created_at: '2026-01-09T11:15:05Z',
    updated_at: '2026-01-09T11:15:05Z'
  },
  {
    block_id: 3,
    ts_start: '2026-01-09T13:00:00Z',
    ts_end: '2026-01-09T14:30:00Z',
    duration_minutes: 90.0,
    duration_hours: 1.5,
    primary_app_name: 'Code.exe',
    primary_domain: null,
    title_summary: 'main.go - ChronicleCore',
    profile_id: 2,
    client_name: 'Internal',
    project_name: 'ChronicleCore',
    service_name: 'Development',
    confidence: 'HIGH',
    billable: false,
    locked: false,
    notes: null,
    description: 'Code - main.go - ChronicleCore',
    created_at: '2026-01-09T14:30:05Z',
    updated_at: '2026-01-09T14:30:05Z'
  }
];

const mockClients = [
  { client_id: 1, name: 'Acme Corp', is_active: true, created_at: '2026-01-09T10:00:00Z', updated_at: '2026-01-09T10:00:00Z' },
  { client_id: 2, name: 'Internal', is_active: true, created_at: '2026-01-09T10:00:00Z', updated_at: '2026-01-09T10:00:00Z' }
];

const mockServices = [
  { service_id: 1, name: 'Bookkeeping', is_active: true },
  { service_id: 2, name: 'Development', is_active: true },
  { service_id: 3, name: 'Consulting', is_active: true }
];

const mockRates = [
  {
    rate_id: 1,
    name: 'Standard',
    currency: 'USD',
    hourly_amount: 150.00,
    hourly_minor_units: 15000,
    effective_from: null,
    effective_to: null,
    is_active: true
  },
  {
    rate_id: 2,
    name: 'Internal',
    currency: 'USD',
    hourly_amount: 0.00,
    hourly_minor_units: 0,
    effective_from: null,
    effective_to: null,
    is_active: true
  }
];

const mockProfiles = [
  {
    profile_id: 1,
    client_name: 'Acme Corp',
    project_name: null,
    service_name: 'Bookkeeping',
    rate_name: 'Standard',
    rate_amount: 150.00,
    currency: 'USD',
    is_active: true
  },
  {
    profile_id: 2,
    client_name: 'Internal',
    project_name: 'ChronicleCore',
    service_name: 'Development',
    rate_name: 'Internal',
    rate_amount: 0.00,
    currency: 'USD',
    is_active: true
  }
];

// Health check
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    version: '1.0.0-mock',
    uptime_seconds: 3600
  });
});

// Tracking endpoints
app.get('/api/v1/tracking/status', (req, res) => {
  res.json({
    state: trackingState,
    last_active_at: lastActiveAt,
    idle_seconds: 0,
    current_window: {
      app_name: 'Code.exe',
      title: 'server.js - Mock API'
    }
  });
});

app.post('/api/v1/tracking/start', (req, res) => {
  trackingState = 'ACTIVE';
  lastActiveAt = new Date().toISOString();
  res.json({
    state: trackingState,
    last_active_at: lastActiveAt,
    idle_seconds: 0,
    current_window: { app_name: 'Code.exe', title: 'server.js - Mock API' }
  });
});

app.post('/api/v1/tracking/pause', (req, res) => {
  if (trackingState !== 'ACTIVE') {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Not active', details: {} } });
  }
  trackingState = 'PAUSED';
  res.json({
    state: trackingState,
    last_active_at: lastActiveAt,
    idle_seconds: 0,
    current_window: null
  });
});

app.post('/api/v1/tracking/resume', (req, res) => {
  if (trackingState !== 'PAUSED') {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Not paused', details: {} } });
  }
  trackingState = 'ACTIVE';
  lastActiveAt = new Date().toISOString();
  res.json({
    state: trackingState,
    last_active_at: lastActiveAt,
    idle_seconds: 0,
    current_window: { app_name: 'Code.exe', title: 'server.js - Mock API' }
  });
});

app.post('/api/v1/tracking/stop', (req, res) => {
  trackingState = 'STOPPED';
  res.json({
    state: trackingState,
    last_active_at: lastActiveAt,
    idle_seconds: 0,
    current_window: null
  });
});

// Blocks endpoints
app.get('/api/v1/blocks', (req, res) => {
  const { unassigned, needs_review, profile_id, date, start_date, end_date, limit } = req.query;

  let filtered = [...mockBlocks];

  if (unassigned === 'true') {
    filtered = filtered.filter(b => b.profile_id === null);
  }

  if (needs_review === 'true') {
    filtered = filtered.filter(b => b.profile_id === null || b.confidence === 'LOW');
  }

  if (profile_id) {
    filtered = filtered.filter(b => b.profile_id === parseInt(profile_id));
  }

  // Date filtering would go here (simplified for mock)

  const limitNum = limit ? parseInt(limit) : 100;
  filtered = filtered.slice(0, limitNum);

  res.json(filtered);
});

app.post('/api/v1/blocks/:id/reassign', (req, res) => {
  const blockId = parseInt(req.params.id);
  const { profile_id, confidence } = req.body;

  const block = mockBlocks.find(b => b.block_id === blockId);
  if (!block) {
    return res.status(404).json({ error: { code: 'NOT_FOUND', message: 'Block not found', details: {} } });
  }

  block.profile_id = profile_id;
  block.confidence = confidence || 'HIGH';

  if (profile_id) {
    const profile = mockProfiles.find(p => p.profile_id === profile_id);
    if (profile) {
      block.client_name = profile.client_name;
      block.service_name = profile.service_name;
      block.project_name = profile.project_name;
    }
  } else {
    block.client_name = null;
    block.service_name = null;
    block.project_name = null;
  }

  block.updated_at = new Date().toISOString();

  res.json(block);
});

app.post('/api/v1/blocks/:id/lock', (req, res) => {
  const blockId = parseInt(req.params.id);
  const { locked } = req.body;

  const block = mockBlocks.find(b => b.block_id === blockId);
  if (!block) {
    return res.status(404).json({ error: { code: 'NOT_FOUND', message: 'Block not found', details: {} } });
  }

  block.locked = locked;
  block.updated_at = new Date().toISOString();

  res.json(block);
});

// Profile management endpoints
app.get('/api/v1/clients', (req, res) => {
  const { active_only } = req.query;
  let filtered = mockClients;
  if (active_only !== 'false') {
    filtered = filtered.filter(c => c.is_active);
  }
  res.json(filtered);
});

app.post('/api/v1/clients/create', (req, res) => {
  const { name } = req.body;
  if (!name) {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Name is required', details: {} } });
  }

  const newClient = {
    client_id: mockClients.length + 1,
    name,
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  };
  mockClients.push(newClient);
  res.status(201).json(newClient);
});

app.get('/api/v1/services', (req, res) => {
  res.json(mockServices);
});

app.post('/api/v1/services/create', (req, res) => {
  const { name } = req.body;
  if (!name) {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Name is required', details: {} } });
  }

  const newService = {
    service_id: mockServices.length + 1,
    name,
    is_active: true
  };
  mockServices.push(newService);
  res.status(201).json(newService);
});

app.get('/api/v1/rates', (req, res) => {
  res.json(mockRates);
});

app.post('/api/v1/rates/create', (req, res) => {
  const { name, currency, hourly_amount } = req.body;
  if (!name || !currency || hourly_amount === undefined) {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Missing required fields', details: {} } });
  }

  const newRate = {
    rate_id: mockRates.length + 1,
    name,
    currency,
    hourly_amount: parseFloat(hourly_amount),
    hourly_minor_units: Math.round(parseFloat(hourly_amount) * 100),
    effective_from: null,
    effective_to: null,
    is_active: true
  };
  mockRates.push(newRate);
  res.status(201).json(newRate);
});

app.get('/api/v1/profiles', (req, res) => {
  res.json(mockProfiles);
});

app.post('/api/v1/profiles', (req, res) => {
  const { client_id, service_id, rate_id, project_id, name } = req.body;

  if (!client_id || !service_id || !rate_id) {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Missing required fields', details: {} } });
  }

  const client = mockClients.find(c => c.client_id === client_id);
  const service = mockServices.find(s => s.service_id === service_id);
  const rate = mockRates.find(r => r.rate_id === rate_id);

  if (!client || !service || !rate) {
    return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Invalid foreign key', details: {} } });
  }

  const newProfile = {
    profile_id: mockProfiles.length + 1,
    client_name: client.name,
    project_name: null,
    service_name: service.name,
    rate_name: rate.name,
    rate_amount: rate.hourly_amount,
    currency: rate.currency,
    is_active: true
  };
  mockProfiles.push(newProfile);
  res.status(201).json(newProfile);
});

app.delete('/api/v1/profiles/:id', (req, res) => {
  const profileId = parseInt(req.params.id);
  const index = mockProfiles.findIndex(p => p.profile_id === profileId);

  if (index === -1) {
    return res.status(404).json({ error: { code: 'NOT_FOUND', message: 'Profile not found', details: {} } });
  }

  mockProfiles[index].is_active = false;
  res.status(204).send();
});

// Export endpoint (simplified)
app.post('/api/v1/export/invoice-lines', (req, res) => {
  const csv = `Client,Project,Service,Date,Start Time,End Time,Hours (Rounded),Hours (Actual),Rate,Currency,Amount,Description,Confidence
Acme Corp,,Bookkeeping,2026-01-09,09:00,10:30,1.50,1.50,150.00,USD,225.00,"Excel - Budget 2026.xlsx",HIGH
Internal,ChronicleCore,Development,2026-01-09,13:00,14:30,1.50,1.50,0.00,USD,0.00,"Code - main.go - ChronicleCore",HIGH`;

  res.setHeader('Content-Type', 'text/csv');
  res.setHeader('Content-Disposition', 'attachment; filename=invoice.csv');
  res.send(csv);
});

app.listen(PORT, '127.0.0.1', () => {
  console.log(`🚀 ChronicleCore Mock API running on http://127.0.0.1:${PORT}`);
  console.log('📋 Available endpoints:');
  console.log('   GET  /health');
  console.log('   GET  /api/v1/tracking/status');
  console.log('   POST /api/v1/tracking/{start,pause,resume,stop}');
  console.log('   GET  /api/v1/blocks');
  console.log('   POST /api/v1/blocks/:id/reassign');
  console.log('   POST /api/v1/blocks/:id/lock');
  console.log('   GET  /api/v1/clients');
  console.log('   POST /api/v1/clients/create');
  console.log('   GET  /api/v1/services');
  console.log('   POST /api/v1/services/create');
  console.log('   GET  /api/v1/rates');
  console.log('   POST /api/v1/rates/create');
  console.log('   GET  /api/v1/profiles');
  console.log('   POST /api/v1/profiles');
  console.log('   DELETE /api/v1/profiles/:id');
  console.log('   POST /api/v1/export/invoice-lines');
});
