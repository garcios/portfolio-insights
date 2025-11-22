#!/bin/bash

# Stop Prometheus & Grafana monitoring stack

echo "🛑 Stopping monitoring stack..."
echo ""

podman stop prometheus grafana alertmanager 2>/dev/null || true
podman rm prometheus grafana alertmanager 2>/dev/null || true

echo "✓ All containers stopped and removed"
echo ""
echo "To remove volumes (deletes all data):"
echo "  podman volume rm prometheus-data grafana-data alertmanager-data"
echo ""
