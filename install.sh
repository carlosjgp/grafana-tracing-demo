#!/bin/sh

NAMESPACE=observability

# Just following Jaerger start up guide
# https://github.com/jaegertracing/jaeger-operator#getting-started
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: observability
EOF
kubectl apply -n $NAMESPACE -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/crds/jaegertracing.io_jaegers_crd.yaml
kubectl apply -n $NAMESPACE -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/service_account.yaml
kubectl apply -n $NAMESPACE -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/role.yaml
kubectl apply -n $NAMESPACE -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/role_binding.yaml
kubectl apply -n $NAMESPACE -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/operator.yaml
kubectl apply -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/cluster_role.yaml
kubectl apply -f https://raw.githubusercontent.com/jaegertracing/jaeger-operator/master/deploy/cluster_role_binding.yaml

kubectl apply -n $NAMESPACE -f - <<EOF
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: simplest
EOF

helm upgrade --install \
    --version 11.6.1 \
    --namespace $NAMESPACE \
    prometheus stable/prometheus \
    -f ./k8s-deployments/prometheus.yaml

helm upgrade --install \
    --version 0.30.1 \
    --namespace observability \
    loki loki/loki \
    -f ./k8s-deployments/loki.yaml

helm upgrade --install \
    --version 0.23.2 \
    --namespace observability \
    promtail loki/promtail \
    -f ./k8s-deployments/promtail.yaml

helm upgrade --install \
    --version 5.3.4 \
    --namespace $NAMESPACE \
    grafana stable/grafana \
    -f ./k8s-deployments/grafana.yaml

kubectl apply --namespace $NAMESPACE -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: grafana-image-renderer
    app.kubernetes.io/part-of: grafana
  name: grafana-image-renderer
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: grafana-image-renderer
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app.kubernetes.io/name: grafana-image-renderer
        app.kubernetes.io/part-of: grafana
      annotations:
        prometheus.io/port: "http"
        prometheus.io/scrape: "true"
    spec:
      containers:
      - image: grafana/grafana-image-renderer:2.0.0
        name: grafana-image-renderer
        resources: {}
        env:
          - name: ENABLE_METRICS
            value: 'true'
        ports:
          - name: http
            containerPort: 8081
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: grafana-image-renderer
    app.kubernetes.io/part-of: grafana
  name: grafana-image-renderer
spec:
  ports:
  - name: http
    port: 8081
    protocol: TCP
    targetPort: http
  selector:
    app.kubernetes.io/name: grafana-image-renderer
  type: ClusterIP
EOF

kubectl apply -f k8s-deployments/observability-demo-app.yaml
