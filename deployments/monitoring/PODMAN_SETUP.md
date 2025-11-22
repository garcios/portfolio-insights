# Monitoring Stack - Podman on macOS Setup

## ✅ Successfully Deployed!

The monitoring stack is now running with Podman on macOS using custom images to work around volume mount limitations.

## 🚀 Quick Commands

```bash
# Start monitoring stack
make monitoring-up

# Stop monitoring stack
make monitoring-down

# View Prometheus logs
make monitoring-logs

# View all logs
podman logs -f prometheus
podman logs -f grafana
podman logs -f alertmanager
```

## 🌐 Access URLs

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/Password123)
- **AlertManager**: http://localhost:9093

## 🔧 How It Works

Due to Podman on macOS limitations with bind mounts, we use a custom approach:

1. **Custom Images**: Configuration files are embedded into custom Docker images
2. **Build Process**: `start-monitoring.sh` builds three custom images:
   - `prometheus-custom` - with prometheus.yml and alerts.yml
   - `grafana-custom` - with provisioning and dashboards
   - `alertmanager-custom` - with alertmanager.yml

3. **No Volume Mounts**: Config files are baked into images, not mounted

## 📝 Updating Configurations

When you update any config files (prometheus.yml, alerts.yml, etc.):

```bash
# Rebuild and restart
make monitoring-down
make monitoring-up
```

This will rebuild the custom images with your new configurations.

## 🐛 Troubleshooting

### Check Container Status
```bash
podman ps | grep -E "prometheus|grafana|alertmanager"
```

### Check Logs
```bash
podman logs prometheus
podman logs grafana
podman logs alertmanager
```

### Restart a Single Service
```bash
podman restart prometheus
podman restart grafana
podman restart alertmanager
```

### Clean Restart
```bash
make monitoring-down
podman volume rm prometheus-data grafana-data alertmanager-data
make monitoring-up
```

## 📊 Verify Metrics Collection

### Check Prometheus Targets
```bash
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

### Check Transaction Service Metrics
```bash
curl http://localhost:9097/metrics | grep grpc_requests_total
```

## 🎯 Next Steps

1. ✅ Monitoring stack is running
2. ⏳ Start transaction-service with metrics enabled
3. ⏳ Create a transaction to generate metrics
4. ⏳ Build Grafana dashboards
5. ⏳ Instrument other services

## 📁 Files

- `deployments/monitoring/start-monitoring.sh` - Start script (builds custom images)
- `deployments/monitoring/stop-monitoring.sh` - Stop script
- `deployments/monitoring/docker-compose.yml` - Original compose file (not used with Podman)
- `Makefile` - Updated with monitoring-up/down targets

## ⚠️ Important Notes

- **Config changes require rebuild**: Since configs are embedded in images, you must rebuild when changing them
- **Data persistence**: Volumes (prometheus-data, grafana-data, alertmanager-data) persist data between restarts
- **Podman-specific**: This setup is optimized for Podman on macOS; Docker users can use docker-compose.yml directly

---

**Status**: ✅ Monitoring stack running successfully with Podman!
