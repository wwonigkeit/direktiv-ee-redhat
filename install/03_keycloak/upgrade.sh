#!/bin/sh

helm upgrade -f db.yaml -f env.yaml -f keycloak_out.yaml -n keycloak --version 18.0.2 keycloak bitnami/keycloak

