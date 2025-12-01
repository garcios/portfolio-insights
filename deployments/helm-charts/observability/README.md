# Observability Helm Chart

This Helm chart deploys the Portfolio Insights observability stack, including Prometheus, Grafana, and Alertmanager.

## Components

*   **Prometheus**: Time-series database for monitoring.
*   **Grafana**: Visualization dashboard.
*   **Alertmanager**: Handles alerts sent by client applications such as Prometheus.

## Prerequisites

*   Kubernetes 1.19+
*   Helm 3.2.0+

## Installation

To install the chart with the release name `observability`:

```bash
helm install observability ./deployments/helm-charts/observability
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
|Str |---|---|
| `global.environment` | Deployment environment (e.g., development, production) | `development` |
| `prometheus.enabled` | Enable Prometheus deployment | `true` |
| `prometheus.image.repository` | Prometheus image repository | `prom/prometheus` |
| `prometheus.image.tag` | Prometheus image tag | `v2.48.0` |
| `prometheus.service.port` | Prometheus service port | `9090` |
| `prometheus.persistence.enabled` | Enable persistence for Prometheus | `true` |
| `prometheus.persistence.size` | Size of Prometheus data volume | `10Gi` |
| `prometheus.config.targets` | Map of service names to their DNS:port for scraping | (see values.yaml) |
| `grafana.enabled` | Enable Grafana deployment | `true` |
| `grafana.image.repository` | Grafana image repository | `grafana/grafana` |
| `grafana.image.tag` | Grafana image tag | `10.2.2` |
| `grafana.service.port` | Grafana service port | `3000` |
| `grafana.admin.user` | Grafana admin username | `admin` |
| `grafana.admin.password` | Grafana admin password | `admin` |
| `grafana.persistence.enabled` | Enable persistence for Grafana | `true` |
| `grafana.persistence.size` | Size of Grafana data volume | `10Gi` |
| `grafana.ingress.enabled` | Enable Ingress for Grafana | `false` |
| `alertmanager.enabled` | Enable Alertmanager deployment | `true` |
| `alertmanager.image.repository` | Alertmanager image repository | `prom/alertmanager` |
| `alertmanager.image.tag` | Alertmanager image tag | `v0.26.0` |
| `alertmanager.service.port` | Alertmanager service port | `9093` |
| `alertmanager.persistence.enabled` | Enable persistence for Alertmanager | `true` |
| `alertmanager.persistence.size` | Size of Alertmanager data volume | `5Gi` |

## Accessing Grafana

1.  **Port Forwarding**:
    ```bash
    kubectl port-forward svc/observability-grafana 3000:3000
    ```
    Access at `http://localhost:3000`.

2.  **Ingress**:
    If ingress is enabled, access at the configured host (default: `grafana.local`).

## Default Credentials

*   **User**: `admin`
*   **Password**: `admin` (Change this in production!)

## Monitoring Targets

The chart is pre-configured to monitor the following services:
*   Gateway Service
*   User Service
*   Transaction Service
*   Portfolio Service
*   Market Data Service

Ensure these services are running and accessible within the cluster. You may need to update the `prometheus.config.targets` in `values.yaml` to match your service DNS names if they differ from the defaults.
