#!/bin/sh

echo "updating linkerd"

helm upgrade linkerd-crds linkerd/linkerd-crds --version 1.8.0 -n linkerd --create-namespace || exit 1

helm upgrade linkerd-control-plane \
  -n linkerd \
  --version 1.15.0 \
  --set-file identityTrustAnchorsPEM=certs/ca.crt \
  --set-file identity.issuer.tls.crtPEM=certs/issuer.crt \
  --set-file identity.issuer.tls.keyPEM=certs/issuer.key \
  linkerd/linkerd-control-plane --wait || exit 1