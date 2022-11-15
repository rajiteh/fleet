#!/bin/bash

set -euxo pipefail

# wait for loadBalancer IPs
{ grep -q -m 1 -e ".*"; kill $!; } < <(kubectl get ingress -n cattle-system rancher -o 'go-template={{range .status.loadBalancer.ingress}}{{.ip}}{{"\n"}}{{end}}' -w)
# wait for certificate
{ grep -q -m 1 -e "tls-rancher-ingress.*True"; kill $!; } < <(kubectl get certs -n cattle-system -w)
