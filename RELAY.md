# Relay
Run `go run ./tools/lab` for ephemeral loopback UDP/TCP E2E (no files or secrets).
Run `go run ./services/relay/cmd/relay -cert server.pem -key server.key -ca devices-ca.pem`.
Listener: UDP :4433. Health and Prometheus metrics: 127.0.0.1:9090.
Every client must have a valid device certificate. No anonymous sessions.
Default destination policy rejects private, loopback, link-local and special
ranges. Prefer an explicit deployment destination allowlist with `-allow-cidrs`.
The lab explicitly permits 127.0.0.1/32 and must never be exposed publicly.
128 active sessions; 64 flows/session; fixed QUIC receive windows; per-session
2000 packet/sec token bucket; UDP idle timeout; bounded client datagram queues.
QUIC authenticated return traffic avoids unauthenticated UDP reflection.
Production DDoS protection, revocation, ticket auth and load telemetry are future phases.
