# Status

## Implemented
- Phase 0 (b74718d): toolchain, four ADRs, protocol and backend boundary.
- Phase 1 (14a93d1): mTLS persistent QUIC sessions, UDP DATAGRAM/TCP streams,
  echo tools, flow/session limits, cleanup, destination policy and metrics.
- Phase 2: DIRECT/RELAY flow handles, route pinning, concurrent forwarding,
  policy changes for new flows, default DIRECT fallback and optional BLOCK.

## Verified on Linux
`go build ./...`, `go test -race -timeout 60s ./...`, `go vet ./...`,
`go run ./tools/lab` all passed after the fallback fix.
Real UDP/TCP E2E, simultaneous DIRECT/RELAY, route pinning, unknown-process
DIRECT, BLOCK fallback, relay-shutdown fallback, mTLS rejection, special-address
policy, UDP idle cleanup and deterministic 50% destination loss tests passed.
Request/reply ordering was checked; UDP transport does not guarantee ordering.
Local RTT measurements are not evidence of Internet route improvement.

## Outstanding
Phase 3: no Windows process discovery, WinDivert/WFP interception, driver or
privileged service. Flow metadata is supplied by the caller in the current API.
Per-process Windows behavior cannot be verified on this Linux worker.
Phases 4–10: multi-relay measurement/scoring/hysteresis, PostgreSQL control plane,
signed config/tickets, WPF GUI, production tuning, full security hardening,
Grafana/alerts and deployment/update pipeline are not implemented.
Only framing microbenchmarks are supplied; no relay throughput claim.
No remote push or deployment was performed. No default routes or adapters changed.

## Parser fuzzing
`FuzzDecode` and `FuzzReadOpen` each ran for 5 seconds with two workers and passed:
31,802 and 29,862 inputs respectively. This is a short smoke fuzz run, not a
complete protocol security audit. Framing microbenchmark results: BENCHMARKS.md.
