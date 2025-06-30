#!/bin/bash 

rm -Rf /tmp/theme
mkdir /tmp/theme
find ../../tests/keycloak/import/keycloak-theme/direktiv -type f -name '*.*' -exec cp {} /tmp/theme \;

kubectl create configmap direktiv-theme -n keycloak --dry-run=client --from-file=/tmp/theme -o yaml > ../03_keycloak/theme.yaml

kubectl create configmap direktiv-import -n keycloak --dry-run=client --from-file=../../tests/keycloak/import/direktiv.json -o yaml > ../03_keycloak/import.yaml
