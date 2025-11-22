#!/bin/bash

# Prometheus & Grafana Monitoring Stack for Podman on macOS
# Works around volume mount limitations by embedding configs in custom images

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Starting Prometheus & Grafana monitoring stack with Podman..."
echo "   (Using custom images with embedded configs for macOS compatibility)"
echo ""

# Build custom Prometheus image with configs
echo "🔨 Building custom Prometheus image..."
cat > Dockerfile.prometheus << 'EOF'
FROM prom/prometheus:v2.48.0
COPY prometheus.yml /etc/prometheus/prometheus.yml
COPY alerts.yml /etc/prometheus/alerts.yml
EOF

cd prometheus
podman build -t prometheus-custom:latest -f ../Dockerfile.prometheus .
cd ..
rm Dockerfile.prometheus
echo "✓ Prometheus image built"
echo ""

# Build custom AlertManager image with config
echo "🔨 Building custom AlertManager image..."
cat > Dockerfile.alertmanager << 'EOF'
FROM prom/alertmanager:v0.26.0
COPY alertmanager.yml /etc/alertmanager/alertmanager.yml
EOF

cd alertmanager
podman build -t alertmanager-custom:latest -f ../Dockerfile.alertmanager .
cd ..
rm Dockerfile.alertmanager
echo "✓ AlertManager image built"
echo ""

# Build custom Grafana image with provisioning
echo "🔨 Building custom Grafana image..."
cat > Dockerfile.grafana << 'EOF'
FROM grafana/grafana:10.2.2
USER root
COPY provisioning /etc/grafana/provisioning
COPY dashboards /var/lib/grafana/dashboards
RUN chown -R 472:472 /etc/grafana/provisioning /var/lib/grafana/dashboards
USER 472
EOF

cd grafana
podman build -t grafana-custom:latest -f ../Dockerfile.grafana .
cd ..
rm Dockerfile.grafana
echo "✓ Grafana image built"
echo ""

# Create network if it doesn't exist
echo "📡 Creating monitoring network..."
podman network exists monitoring 2>/dev/null || podman network create monitoring
echo "✓ Network ready"
echo ""

# Create volumes
echo "💾 Creating volumes..."
podman volume exists prometheus-data 2>/dev/null || podman volume create prometheus-data
podman volume exists grafana-data 2>/dev/null || podman volume create grafana-data
podman volume exists alertmanager-data 2>/dev/null || podman volume create alertmanager-data
echo "✓ Volumes ready"
echo ""

# Stop and remove existing containers if they exist
echo "🧹 Cleaning up existing containers..."
podman rm -f prometheus grafana alertmanager 2>/dev/null || true
echo "✓ Cleanup complete"
echo ""

# Start Prometheus
echo "🔥 Starting Prometheus..."
podman run -d \
  --name prometheus \
  --network monitoring \
  -p 9090:9090 \
  -v prometheus-data:/prometheus \
  prometheus-custom:latest \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --web.console.libraries=/usr/share/prometheus/console_libraries \
  --web.console.templates=/usr/share/prometheus/consoles \
  --web.enable-lifecycle

echo "✓ Prometheus started on http://localhost:9090"
echo ""

# Start Grafana
echo "📊 Starting Grafana..."
podman run -d \
  --name grafana \
  --network monitoring \
  -p 3001:3000 \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  -e GF_USERS_ALLOW_SIGN_UP=false \
  -e GF_SERVER_ROOT_URL=http://localhost:3001 \
  -v grafana-data:/var/lib/grafana \
  grafana-custom:latest

echo "✓ Grafana started on http://localhost:3001 (admin/admin)"
echo ""

# Start AlertManager
echo "🔔 Starting AlertManager..."
podman run -d \
  --name alertmanager \
  --network monitoring \
  -p 9093:9093 \
  -v alertmanager-data:/alertmanager \
  alertmanager-custom:latest \
  --config.file=/etc/alertmanager/alertmanager.yml \
  --storage.path=/alertmanager

echo "✓ AlertManager started on http://localhost:9093"
echo ""

echo "========================================="
echo "✅ Monitoring stack is running!"
echo "========================================="
echo ""
echo "Access the services:"
echo "  Prometheus:   http://localhost:9090"
echo "  Grafana:      http://localhost:3001 (admin/admin)"
echo "  AlertManager: http://localhost:9093"
echo ""
echo "To view logs:"
echo "  podman logs -f prometheus"
echo "  podman logs -f grafana"
echo "  podman logs -f alertmanager"
echo ""
echo "To stop all services:"
echo "  ./stop-monitoring.sh"
echo ""
echo "Note: Config files are embedded in custom images."
echo "To update configs, re-run this script to rebuild images."
echo ""
