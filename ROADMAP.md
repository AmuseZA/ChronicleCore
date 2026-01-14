# ChronicleCore Product Roadmap

**Current Version:** 1.8.1  
**Last Updated:** January 2026

---

## Vision

Transform ChronicleCore from a powerful time tracker into an **intelligent productivity platform** that automatically categorizes work, generates insights, and integrates with billing/invoicing workflows.

---

## 🎯 Strategic Priorities

1. **ML Intelligence** – Make automatic categorization smarter and more accurate
2. **Billing Integration** – Connect tracked time to invoicing workflows
3. **Automation** – Reduce manual review through smart rules
4. **User Experience** – Streamline the daily workflow
5. **Extensibility** – Enable integrations with external tools

---

## 📅 Release Timeline

### v1.9.0 – ML Intelligence (Q1 2026)
**Theme:** Make the ML smarter and more actionable

| Feature | Description | Priority |
|---------|-------------|----------|
| **Confidence Thresholds** | Auto-assign profiles when ML confidence > 90% | High |
| **Training Feedback Loop** | "Accept/Reject" buttons train the model in real-time | High |
| **Multi-label Support** | Allow blocks to have multiple tags/categories | Medium |
| **ML Insights Dashboard** | Show accuracy stats, training samples, model health | Medium |
| **Scheduled Training** | Auto-retrain model nightly with new data | Low |

---

### v1.10.0 – Reporting & Exports (Q1 2026)
**Theme:** Professional reports and flexible data export

| Feature | Description | Priority |
|---------|-------------|----------|
| **PDF Report Generation** | Branded time reports for clients | High |
| **Date Range Summaries** | Weekly/monthly summary views | High |
| **CSV/Excel Export** | Filtered exports for accounting | High |
| **Client-based Grouping** | Group time by client across profiles | Medium |
| **Scheduled Email Reports** | Auto-send weekly summaries to self/clients | Low |

---

### v1.11.0 – Automation Rules (Q2 2026)
**Theme:** Reduce manual review through smart automation

| Feature | Description | Priority |
|---------|-------------|----------|
| **Rule Builder UI** | Visual interface to create classification rules | High |
| **Time-based Rules** | "After 6pm = Personal" type rules | High |
| **Combined Conditions** | App + Keyword + Time compound rules | Medium |
| **Rule Priority System** | Define which rules take precedence | Medium |
| **Rule Templates** | Pre-built rules for common use cases | Low |

---

### v2.0.0 – Invoicing Integration (Q2 2026)
**Theme:** Connect time tracking to billing

| Feature | Description | Priority |
|---------|-------------|----------|
| **Invoice Generation** | Create invoices from tracked time | High |
| **Xero Integration** | Push time entries to Xero as line items | High |
| **Stripe Connect** | Accept payments directly | Medium |
| **QuickBooks Integration** | Alternative accounting integration | Medium |
| **Recurring Invoices** | Auto-generate monthly invoices per client | Low |

---

### v2.1.0 – UI Polish & Performance (Q3 2026)
**Theme:** Streamline the daily experience

| Feature | Description | Priority |
|---------|-------------|----------|
| **Dashboard Widgets** | Customizable dashboard with key metrics | High |
| **Quick Actions** | Keyboard shortcuts for common tasks | High |
| **Activity Timeline** | Visual timeline view of the day | Medium |
| **Dark Mode** | System-aware dark theme | Medium |
| **Performance Mode** | Reduced UI updates for low-power systems | Low |

---

### v2.2.0 – Integrations & API (Q3 2026)
**Theme:** Connect with the broader ecosystem

| Feature | Description | Priority |
|---------|-------------|----------|
| **Public REST API** | Documented API for third-party tools | High |
| **Calendar Sync** | Import Google/Outlook calendar events | High |
| **Slack/Teams Notifications** | Daily summary notifications | Medium |
| **Zapier/Make Webhooks** | Trigger automations on events | Medium |
| **Browser Extension** | Quick timer start/stop from browser | Low |

---

## 🔮 Future Considerations (v3.0+)

| Concept | Description |
|---------|-------------|
| **Mobile App** | iOS/Android companion for on-the-go entries |
| **Team Features** | Multi-user support with shared clients |
| **AI Assistant** | Natural language queries ("How much time on Project X?") |
| **Offline Mode** | Full functionality without internet |
| **Plugin System** | Community-contributed integrations |

---

## 🛠️ Technical Debt / Infrastructure

| Item | Version | Notes |
|------|---------|-------|
| Update SvelteKit to latest | v1.9.0 | Fix `selectedId` lint errors |
| Add A11y attributes to modals | v1.9.0 | Address existing warnings |
| Database migrations versioning | v1.10.0 | Track schema changes properly |
| Automated UI tests | v2.0.0 | Playwright integration |
| CI/CD pipeline | v2.0.0 | GitHub Actions for builds |

---

## 📊 Success Metrics

| Metric | Target |
|--------|--------|
| ML auto-assignment accuracy | > 85% |
| Time to review daily activities | < 5 minutes |
| Invoice generation time | < 30 seconds |
| User-reported bugs per release | < 3 |

---

## How to Contribute Ideas

Create a GitHub issue with the **`enhancement`** label, or discuss in the project's Discussions tab.
