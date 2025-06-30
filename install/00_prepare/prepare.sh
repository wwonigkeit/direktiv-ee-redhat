#!/bin/sh
set -eu

cd "$(dirname "$0")"

# Ensure the working directory for generated certs
mkdir -p linkerd

echo "Generating certs using Go..."
go run main.go

# Replace the existing certs in the Linkerd folder
rm -rf ../01_linkerd/certs
mv linkerd ../01_linkerd/certs

# Add required Helm repositories
helm repo add linkerd https://helm.linkerd.io/stable
helm repo add percona https://percona.github.io/percona-helm-charts/
helm repo add direktiv https://charts.direktiv.io
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add nginx https://kubernetes.github.io/ingress-nginx
helm repo add prometheus https://prometheus-community.github.io/helm-charts
helm repo add fluent-bit https://fluent.github.io/helm-charts
helm repo add nats https://nats-io.github.io/k8s/helm/charts
helm repo add opensearch https://opensearch-project.github.io/helm-charts/
helm repo add opentelemetry-collector https://open-telemetry.github.io/opentelemetry-helm-charts

helm repo update
