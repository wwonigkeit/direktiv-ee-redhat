#!/bin/sh
cd "$(dirname "$0")"

helm install --create-namespace -n postgres --version 2.4.1 pg-operator percona/pg-operator --wait

kubectl apply -f database.yaml
