## Description
## Type of Change
- [ ] 🚀 New Feature
- [ ] 🐛 Bug Fix
- [ ] 🧹 Refactor / Tech Debt
- [ ] 📈 Observability / Monitoring Update

---

## 🛠 Engineering Checklist

### 1. Logic & Functionality
- [ ] **Acceptance Criteria:** Does this satisfy the requirements of the ticket?
- [ ] **Edge Cases:** Have you handled nulls, empty states, and boundary values?
- [ ] **Backward Compatibility:** Does this change break existing clients or DB schemas?
- [ ] **Tests:** Are unit and integration tests included and passing?

### 2. Readability & Maintainability
- [ ] **Self-Documenting:** Are names clear and intent-based?
- [ ] **The "Why":** Are complex logic blocks explained with comments (not just the "what")?
- [ ] **Feature Flags:** If high-risk, is this wrapped in a toggle/feature flag?
- [ ] **DRY:** Have you avoided unnecessary code duplication?

### 3. Performance & Resources
- [ ] **Complexity:** Are algorithms efficient (avoiding $O(n^2)$ where possible)?
- [ ] **Database:** Have you checked for N+1 queries and ensured proper indexing?
- [ ] **Leak Prevention:** Are all connections, streams, and observers properly closed?

### 4. Observability & Logging (The "Ops" Layer)
- [ ] **Structured Logs:** Are logs emitted in JSON format for the log aggregator?
- [ ] **Log Levels:** Are you using `DEBUG`/`INFO`/`WARN`/`ERROR` appropriately?
- [ ] **Data Hygiene:** Confirmed **zero** PII, tokens, or secrets are leaked in logs.
- [ ] **Telemetry:** Are new metrics or "Golden Signals" (Latency, Errors, etc.) instrumented?
- [ ] **Traceability:** Is the `Trace-ID`/`Correlation-ID` preserved through the flow?

### 5. Security & Resilience
- [ ] **Sanitization:** Is all external input validated and sanitized?
- [ ] **Circuit Breakers:** Do external calls have timeouts and fail-fast logic?
- [ ] **Retries:** Does network logic use exponential backoff (not immediate loops)?
- [ ] **Secrets:** Confirmed no credentials or keys are hardcoded.

---

## 🛡️ The Production Hardening Checklist (The "Ops" Layer)

### 1. Deployment & Configuration
- [ ] **Config Separation:** Are we using environment variables (e.g., `.env`, `k8s secrets`) and avoiding hardcoded values?
- [ ] **Health Checks:** Is there a `/healthz` endpoint for the orchestrator (K8s/ECS) to check?
- [ ] **Resource Limits:** Are CPU/Memory requests and limits defined for the container/pod?
- [ ] **Rolling Updates:** Is the deployment strategy set to `RollingUpdate` to ensure zero downtime?

### 2. Monitoring & Alerting
- [ ] **Alert Thresholds:** Are alerts configured for the "Golden Signals" (e.g., "Alert if P95 latency > 500ms for 5 minutes")?
- [ ] **Dashboard Coverage:** Does the Grafana dashboard show the current state of the service?
- [ ] **Log Retention:** Is there a policy for how long logs are kept in the central system?

### 3. Data Integrity & Recovery
- [ ] **Backups:** Is the database backed up regularly? (If managed service, is it enabled?)
- [ ] **Schema Migrations:** Are database migrations idempotent (safe to run multiple times) and versioned?
- [ ] **Data Validation:** Are there "Dead Letter Queues" (DLQs) for failed messages that couldn't be processed?

### 4. Security & Compliance
- [ ] **Network Policies:** Are firewalls/security groups configured to only allow necessary traffic?
- [ ] **Audit Logs:** Are critical actions (e.g., "User deleted", "Admin changed settings") logged with user context?
- [ ] **Dependency Scanning:** Has `npm audit`, `go mod tidy`, or equivalent been run recently?

---

## 🎯 The "Golden Ticket" (The "What")

This is the single most important question. If you can answer this, you've done 80% of the job.

### 1. The "Why" Check
- [ ] **Problem Statement:** Can I articulate the business problem this code solves in one sentence?
- [ ] **User Impact:** How does this change the experience for the end-user?
- [ ] **Success Metric:** How will we know this change was successful (e.g., "Reduce load time by 200ms", "Increase conversion by 5%")?

### 2. The "How" Check (The Implementation)
- [ ] **Requirements Met:** Does the code meet all explicit requirements from the ticket?
- [ ] **Unintended Consequences:** Does this change negatively impact any other part of the system?
- [ ] **Future-Proofing:** Is the code flexible enough to handle the next logical extension of this feature?

---

