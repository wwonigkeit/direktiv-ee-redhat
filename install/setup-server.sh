#!/bin/bash

echo ">>> update packages >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
sudo dnf update -y
sudo dnf install -y jq golang unzip make curl bash-completion

echo ">>> install k3s >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
curl -sfL https://get.k3s.io | sh -s - --disable traefik --write-kubeconfig-mode=644 --cluster-init

# Configure kubectl alias and completion
alias kc="kubectl"
if [ -f /usr/share/bash-completion/completions/kubectl ]; then
    source /usr/share/bash-completion/completions/kubectl
else
    source <(kubectl completion bash)
fi
complete -F __start_kubectl kc
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

cat << 'EOF' >> ~/.bashrc

# Add the KUBECONFIG variable for kubectl
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
EOF

echo ">>> install docker >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
# Install Docker using the official script
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Start and enable Docker
sudo systemctl enable --now docker

# Run local Docker registry
docker run -d -p 5000:5000 --restart=always --name registry registry:2

echo ">>> install helm >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
HELM_VERSION="v3.18.3"
curl -LO https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz
tar -zxvf helm-${HELM_VERSION}-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/local/bin/helm
rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# Add Helm repositories
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
