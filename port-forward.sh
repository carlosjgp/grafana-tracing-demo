#!/bin/sh

pkill kubectl

nohup kubectl port-forward -n observability svc/prometheus-server 9090:80 >/dev/null 2>&1 &
nohup kubectl port-forward -n observability svc/simplest-query 16686 >/dev/null 2>&1 &
nohup kubectl port-forward -n observability svc/loki 3100 >/dev/null 2>&1 &
nohup kubectl port-forward -n observability svc/grafana 3000:80 >/dev/null 2>&1 &

echo "Prometheus:   http://localhost:9090"
echo "Jaeger:       http://localhost:16686"
echo "Loki:         http://localhost:3100"
echo "Grafana:      http://localhost:3000"

echo "Run 'pkill kubectl' to clean up"