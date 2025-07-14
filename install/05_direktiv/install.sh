#!/bin/sh
cd "$(dirname "$0")"

: "${CHARTS_VERSION:=0.9.1}" # Default to 0.9.1 if not already set

# Wait for PostgreSQL secret to exist
while ! kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv >/dev/null 2>&1; do
  echo "postgres not ready, waiting..."
  sleep 5
done

# Create TLS secret if not already present
if ! kubectl get secret direktiv-tls >/dev/null 2>&1; then
  kubectl create secret tls direktiv-tls \
    --cert=../00_prepare/server.crt \
    --key=../00_prepare/server.key
fi

# Generate db.yaml from PostgreSQL secret
tee db.yaml <<EOF
database:
  host: $(kubectl get secret -n postgres direktiv-cluster-pguser-direktiv -o go-template='{{ index .data "host" | base64decode }}')
  port: $(kubectl get secret -n postgres direktiv-cluster-pguser-direktiv -o go-template='{{ index .data "port" | base64decode }}')
  user: $(kubectl get secret -n postgres direktiv-cluster-pguser-direktiv -o go-template='{{ index .data "user" | base64decode }}')
  password: "$(kubectl get secret -n postgres direktiv-cluster-pguser-direktiv -o go-template='{{ index .data "password" | base64decode }}')"
  name: $(kubectl get secret -n postgres direktiv-cluster-pguser-direktiv -o go-template='{{ index .data "dbname" | base64decode }}')
EOF

# Replace placeholder in YAML template
baseTemplate="direktiv.yaml"
rm -f direktiv_out.yaml
sed "s|DIREKTIV_HOST|${DIREKTIV_HOST}|" "$baseTemplate" > direktiv_out.yaml

# Install Direktiv Helm chart
baseCharts="direktiv/direktiv"
helm install -f db.yaml -f direktiv_out.yaml -f keys.yaml direktiv "$baseCharts" --version "${CHARTS_VERSION}"

# Optional: for loading a local Helm chart instead
# cd ../../../charts/direktiv && helm dependency update && helm dependency build
# baseCharts="../../../charts/direktiv"
# helm install -f db.yaml -f direktiv_out.yaml -f keys.yaml direktiv "$baseCharts"

# Clean up legacy namespace if it exists
if kubectl get ns contour-external >/dev/null 2>&1; then
  kubectl delete ns contour-external
fi
