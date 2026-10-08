# neexitlag
Application-aware DIRECT/RELAY route optimizer, built in staged increments.
Current verified phases and limitations are recorded in docs/STATUS.md.

## Development
Go 1.25+ on Linux or Windows. `go build ./...`, `go test -race ./...`,
`go vet ./...` (or `make verify` on Linux). Linux race checks require a C compiler.
No global route or interface changes. Packet interception is a separate Windows phase.
See RELAY.md and PROTOCOL.md. This repository is not yet a deployable Windows optimizer.

## Implemented data plane
`go run ./tools/lab` starts ephemeral mTLS identities, a QUIC relay and UDP/TCP
loopback echo targets, then prints genuine E2E results. For a persistent relay,
see RELAY.md; certificates must be provided explicitly.

`core/routing.Engine` opens DIRECT or RELAY flow handles, pins their route,
limits active flows, and applies policy changes to new flows. Failed relay opens
fall back to DIRECT unless BLOCK was explicitly selected. Unknown process identity
always uses DIRECT. Tests run both routes concurrently and verify relay counters.
The caller currently supplies process metadata; Windows discovery, executable
profile matching, transparent capture and the privileged service are not delivered.
Applications cannot be optimized transparently until Phase 3 is implemented.
