# neexitlag
Application-aware DIRECT/RELAY route optimizer, built in staged increments.
Current verified phases and limitations are recorded in docs/STATUS.md.

## Development
Go 1.25+ on Linux or Windows. `go build ./...`, `go test -race ./...`,
`go vet ./...` (or `make verify` on Linux). Linux race checks require a C compiler.
No global route or interface changes. Packet interception is a separate Windows phase.
See RELAY.md and PROTOCOL.md. This repository is not yet a deployable Windows optimizer.
