#!/bin/sh
cd "$(dirname "$0")"

set -e

while ! kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv
do
  echo "postgres not ready, waiting..."
  sleep 5
done

if ! kubectl get ns keycloak ; then
    kubectl create ns keycloak
fi

if ! kubectl get secrets -n keycloak direktiv-tls ; then
    kubectl create secret tls --cert=../00_prepare/server.crt --key=../00_prepare/server.key direktiv-tls -n keycloak
fi

if [ ! -n "${DIREKTIV_HOST+1}" ]; then
    echo "DIREKTIV_HOST not set to hostname"
    exit 1
fi 

sed "s/DIREKTIV_HOST/`echo $DIREKTIV_HOST`/g" keycloak.yaml > keycloak_out.yaml

tee env.yaml <<EOF
extraEnvVars:
  - name: KC_HOSTNAME_PATH
    value: "/auth/"
  - name: KC_HOSTNAME_STRICT
    value: "false"
  - name: KC_HOSTNAME_ADMIN_URL
    value: "https://${DIREKTIV_HOST}/auth"
  - name: KC_HOSTNAME_URL
    value: "https://${DIREKTIV_HOST}/auth"
  - name: KEYCLOAK_EXTRA_ARGS
    value: "--import-realm"
  - name: KEYCLOAK_PRODUCTION
    value: "true"
  - name: KC_LOG_LEVEL
    value: "DEBUG"
EOF

tee db.yaml <<EOF
externalDatabase:
  host: $(kubectl get secret -n postgres direktiv-cluster-pguser-keycloak -o go-template='{{ index .data "host" | base64decode }}')
  port: $(kubectl get secret -n postgres direktiv-cluster-pguser-keycloak -o go-template='{{ index .data "port" | base64decode }}')
  user: $(kubectl get secret -n postgres direktiv-cluster-pguser-keycloak -o go-template='{{ index .data "user" | base64decode }}')
  password: "$(kubectl get secret -n postgres direktiv-cluster-pguser-keycloak -o go-template='{{ index .data "password" | base64decode }}')"
  database: $(kubectl get secret -n postgres direktiv-cluster-pguser-keycloak -o go-template='{{ index .data "dbname" | base64decode }}')
EOF


sed "s/localhost:8080/`echo $DIREKTIV_HOST`/g" import.yaml > import_out.yaml

kubectl apply -f import_out.yaml
kubectl apply -f theme.yaml

helm install -f db.yaml -f env.yaml -f keycloak_out.yaml -n keycloak --version 18.0.2 keycloak bitnami/keycloak
