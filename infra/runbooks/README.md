# Runbooks

Every service and scheduled job should have a runbook before it's considered “done”.

## Template

Copy this template into a new file per service/job (e.g. `gateway.md`, `md.md`, `daily_mtm.md`).

### Symptoms

- What the user/operator notices (timeouts, wrong numbers, missing updates).

### Checks

- Dashboards to check (Prometheus/Grafana panels, logs, traces).
- Commands to run (curl endpoints, DB queries, health checks).

### Remediation

- Safe steps to mitigate (restart service, clear cache keys, re-run job idempotently).

### Escalation

- What evidence to collect (logs, trace IDs, metrics screenshot).
- Who/what to page (future multi-tenant ops; in v1 this is “you”).
