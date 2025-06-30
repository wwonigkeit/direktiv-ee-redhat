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
# Remove old version of Docker:
dnf remove docker \
                  docker-client \
                  docker-client-latest \
                  docker-common \
                  docker-latest \
                  docker-latest-logrotate \
                  docker-logrotate \
                  docker-engine \
                  podman \
                  runc
# Install Docker using the official script
dnf -y install dnf-plugins-core
dnf config-manager --add-repo https://download.docker.com/linux/rhel/docker-ce.repo
sed -i -e 's/\$releasever/9/g' /etc/yum.repos.d/docker-ce.repo
dnf install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Start and enable Docker
systemctl enable --now docker

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

# Allow pod network CIDR for inter-pod communication
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
sudo firewall-cmd --permanent --add-port=6443/tcp                           # Kubernetes API server
sudo firewall-cmd --permanent --add-port=10250/tcp                          # Kubelet API
sudo firewall-cmd --permanent --add-port=8472/udp                           # VXLAN (Flannel)
sudo firewall-cmd --permanent --add-port=2379-2380/tcp                      # etcd client/server
sudo firewall-cmd --reload
