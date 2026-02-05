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

### 5.