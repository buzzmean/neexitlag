# Status
Phase 0 completed (b74718d): toolchain, ADRs, protocol and backend boundary.
Phase 1 completed: authenticated persistent QUIC session, UDP DATAGRAM/TCP streams,
echo tools, limits, cleanup, destination policy, health and Prometheus counters.
Verified: `go build ./...`, `go test -race -timeout 60s ./...`, `go vet ./...`,
`go run ./tools/lab`. UDP and TCP E2E passed; auth/policy/cleanup/shutdown tests passed.
Local loopback measurements are not evidence of Internet route improvement.
Windows interception, control plane and desktop GUI are not implemented.
