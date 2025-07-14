#!/bin/sh
cd "$(dirname "$0")"

kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.12.2/operator.yaml

kubectl create ns knative-serving

# for proxy use the proxy template
kubectl apply -f basic.yaml

kubectl apply --filename https://github.com/knative/net-contour/releases/download/knative-v1.11.0/contour.yaml
