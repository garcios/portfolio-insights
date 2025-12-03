# Post-Mortem: Prometheus Metrics Display Issue in Grafana

**Date:** 2025-12-03
**Status:** Resolved
**Severity:** Medium (Observability Impact)

## Issue Summary
Users reported that while Prometheus metrics were allegedly working (endpoints accessible), they were not being displayed in the Grafana dashboard. The dashboard panels showed "No data" or empty graphs.

## Root Cause Analysis
Investigation revealed two distinct root causes contributing to the issue:

1.  **Datasource UID Mismatch**:
    *   The Grafana dashboard JSON was hardcoded to use a datasource with `uid: "prometheus"`.
    *   The datasource provisioning configuration (`datasources.yml`) defined a Prometheus datasource but did not explicitly assign a `uid`. Consequently, Grafana generated a random UID or used a default, causing the dashboard panels to fail to find the correct datasource.

2.  **Metric Name Mismatch**:
    *   The dashboard queries referenced generic metric names: `http_requests_total` and `users_created_total`.
    *   The actual metrics exposed by the services were namespaced:
        *   Gateway Service: `gateway_http_requests_total`
        *   User Service: `user_users_created_total`
    *   This discrepancy resulted in empty query results for the affected panels.

## Resolution
The following steps were taken to resolve the issue:

1.  **Configuration Update**:
    *   Modified `deployments/monitoring/grafana/provisioning/datasources/prometheus.yml` to explicitly set `uid: prometheus`. This ensured the datasource created by provisioning matched the UID expected by the dashboard.

2.  **Dashboard Correction**:
    *   Updated `deployments/monitoring/grafana/dashboards/portfolio-insights-dashboard.json` to use the correct, namespaced metric names:
        *   Replaced `rate(http_requests_total[5m])` with `rate(gateway_http_requests_total[5m])`.
        *   Replaced `sum(users_created_total)` with `sum(user_users_created_total)`.

3.  **Deployment**:
    *   Restarted the monitoring stack (`make monitoring-down` && `make monitoring-up`) to rebuild the custom images with the updated configurations and reload the services.

## Verification
*   **Prometheus Targets**: Verified that all service targets (Gateway, User, Transaction, Portfolio, MarketData, Login-Consent) are in `UP` state via the Prometheus API.
*   **Metric Existence**: Verified the existence of the correct metric names using `curl` against the service metrics endpoints.
*   **Grafana Dashboard**: Confirmed that the dashboard now displays data for all panels.

## Lessons Learned & Action Items
*   **Explicit UIDs**: Always explicitly define `uid` in Grafana datasource provisioning files to prevent mismatches with dashboards.
*   **Metric Verification**: Verify exact metric names (including namespaces) using `curl` or Prometheus Expression Browser before building or importing dashboards.
*   **Namespace Awareness**: Be aware that some Prometheus client libraries or middleware may automatically prefix metrics with the service name or a configured namespace.
