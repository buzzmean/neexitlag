#!/usr/bin/env bash
set -euo pipefail
# systemd supplies settings; never source an env file as executable shell code.
allow=${RELAY_ALLOW_CIDRS:-${VPS_PUBLIC_IP:+$VPS_PUBLIC_IP/32}}
[[ -n $allow ]] || { echo 'Explicit relay destination CIDRs required.' >&2; exit 1; }
exec /opt/neexitlag/bin/relay \
 -listen "${RELAY_LISTEN:-:4433}" -metrics "${RELAY_METRICS:-127.0.0.1:9090}" \
 -cert /etc/neexitlag/relay/server.pem -key /etc/neexitlag/relay/server.key \
 -ca /etc/neexitlag/relay/ca.pem -relay-id "${RELAY_ID:-VPS-01}" \
 -ticket-keys /etc/neexitlag/relay/trust.json -replay-cache /var/lib/neexitlag/replay.json \
 -control-plane "${RELAY_CONTROL_PLANE:-https://127.0.0.1:8443}" \
 -control-plane-ca "${RELAY_CONTROL_PLANE_CA_FILE:-/etc/neexitlag/relay/ca.pem}" \
 -heartbeat-token-file /etc/neexitlag/relay/heartbeat.token -allow-cidrs "$allow"
