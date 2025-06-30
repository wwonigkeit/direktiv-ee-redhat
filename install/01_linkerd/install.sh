#!/bin/sh
set -eu

cd "$(dirname "$0")"

# Ensure helm and kubectl are available
command -v helm >/dev/null 2>&1 || { echo >&2 "Helm not found. Please install Helm."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo >&2 "kubectl not found. Please install kubectl."; exit 1; }

# Install Linkerd CRDs
echo "Installing Linkerd CRDs..."
helm install linkerd-crds linkerd/linkerd-crds \
  --version 1.8.0 \
  -n linkerd \
  --create-namespace || exit 1

# Install Linkerd Control Plane with certs
echo "Installing Linkerd control plane..."
helm install linkerd-control-plane \
  linkerd/linkerd-control-plane \
  --version 1.16.9 \
  -n linkerd \
  --set-file identityTrustAnchorsPEM=certs/ca.crt \
  --set-file identity.issuer.tls.crtPEM=certs/issuer.crt \
  --set-file identity.issuer.tls.keyPEM=certs/issuer.key \
  --set proxy.resources.cpu.limit=150m \
  --set proxy.resources.cpu.request=100m \
  --wait || exit 1

# Enable Linkerd injection in target namespaces
for ns in keycloak default; do
  if ! kubectl get ns "$ns" >/dev/null 2>&1; then
    echo "Creating namespace: $ns"
    kubectl create ns "$ns"
  fi

  echo "Annotating namespace $ns for Linkerd injection"
  kubectl annotate ns "$ns" linkerd.io/inject=enabled --overwrite
done
