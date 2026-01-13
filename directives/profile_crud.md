# Directive D07: Profile CRUD and API

**Goal:** Implement the "Settings" data usage: Clients, Projects, Services, Rates.

**Scope:**
- `chroniclecore/internal/api`
- `chroniclecore/internal/store`

**Inputs:**
- `profile`, `client`, `service`, `rate` tables.

**Outputs:**
- REST Endpoints for frontend management.

## 1. Data Access Layer
- Implement CRUD (Create, Read, Update, Delete) methods for:
  - `Client`
  - `Profile` (The join entity)
  - `Rule`

## 2. API Endpoints
- **GET /api/v1/profiles**: List all active profiles (Joined with Client/Service names).
- **POST /api/v1/profiles**: Create new.
- **GET /api/v1/rules**: List rules.
- **POST /api/v1/rules**: Create rule.
- **POST /api/v1/rules/test**: Body `{"app":"foo", "title":"bar"}` -> Returns `{"match": true, "rule_id": 1}`.

## Acceptance Criteria
- [ ] Can create a "Client A".
- [ ] Can create a "Standard Rate".
- [ ] Can link them in a "Profile".
- [ ] JSON response is clean (no internal DB nulls).
