#!/bin/bash
cd "$(dirname "$0")"

if [ -z "${DIREKTIV_HOST}" ]; then
    echo "The environment variable 'DIREKTIV_HOST' is not set or is empty."
    exit 1
fi

# Ensure helm is installed and the repo is initialized
command -v helm >/dev/null 2>&1 || {
    echo "Helm is not installed. Please install Helm before continuing."
    exit 1
}
helm repo update

# Set up kubectl alias and completion
alias kc="kubectl"
if [ -f /usr/share/bash-completion/completions/kubectl ]; then
    source /usr/share/bash-completion/completions/kubectl
elif command -v kubectl &>/dev/null; then
    source <(kubectl completion bash)
fi
complete -F __start_kubectl kc

# Set KUBECONFIG path (adjust if needed for RHEL-based Kubernetes setup)
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

echo ">>> Step 00_prepare >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
./00_prepare/prepare.sh

echo ">>> Step 01_linkerd >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
./01_linkerd/install.sh

echo ">>> Step 02_postgres >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
./02_postgres/install.sh

echo ">>> Step 03_keycloak >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
./03_keycloak/install.sh

echo ">>> Step 04_direktiv >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
./04_direktiv/install.sh
