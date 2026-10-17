# VELIQ

VELIQ is an application-aware Windows DIRECT/RELAY route optimizer. It measures
DIRECT and authenticated QUIC relay paths, ranks their RTT, jitter, loss, failures,
load and stability, and changes the route of **new flows** only when improvement
survives hysteresis. Better Internet latency depends on your destination and relay
placement; this project makes no throughput or Internet improvement claim.

## Нативный Windows-клиент

Русский WPF-клиент запускается отдельным `VELIQ.exe` без прав администратора.
Браузер не нужен; служба продолжает выполнять перехват и маршрутизацию.

```powershell
.\tools\windows\build-desktop.ps1
.\bin\release\VELIQ.exe
```

.NET SDK 8 требуется для сборки; опубликованный EXE включает runtime.
[Интерфейс, IPC, сборка и ограничения](apps/desktop/README.md).
Старый web demo предназначен для разработки и тестов.

## Architecture

```text
WPF desktop → authenticated named pipe → Windows service → routing core
                                                        → WinDivert capture
Client → signed config / device proof / tickets → PostgreSQL control plane
Client → mTLS + ticket → QUIC relay → UDP/TCP destination
```

The control plane never forwards application packets. Existing TCP connections
stay pinned. Unknown processes and unsupported capture traffic stay DIRECT.
No default route, adapter or routing table is changed.

## Current status

Phase 0–2 checkpoint `7967f8352e810d39dea6c793b0bf8cfa20d027f8` is preserved.
Multi-relay selection, real probing, PostgreSQL APIs, signed tickets/configs,
profiles, service IPC, WPF source and a portable demo are implemented.
The external WinDivert v2 integration captures **IPv4 UDP and TCP**. The owner has
reported successful real Windows UDP E2E, simultaneous per-process routing and
service-crash fail-open. New TCP/browser runtime acceptance remains separate and
pending. IPv6 relay transport and the production WFP driver remain unsupported.
See [TCP/web implementation and exact acceptance commands](docs/WINDOWS_TCP_WEB.md).
See [status](docs/STATUS.md), [Windows acceptance](docs/BLOCKERS.md) and
[security](docs/SECURITY.md).

## Real Windows / remote relay E2E

The next acceptance stage is prepared, **READY FOR VPS**. No real Windows/public
Internet capture PASS is claimed. See [exact setup and acceptance commands](docs/WINDOWS_REAL_E2E.md)
and [official pinned WinDivert dependency](docs/WINDIVERT_DEPENDENCY.md).
`tools/windows/setup-e2e.ps1` provisions the remote backend and Windows service;
`acceptance.ps1` checks two actual processes, route evidence, relay failure and
service stop/crash. A local demo or cross-build does not satisfy that acceptance.

## Existing VPS relay benchmark

[First real relay test](docs/FIRST_REAL_RELAY_TEST.md) documents secure relay
install/update/uninstall on the existing VPS and a standalone Windows/Linux
`tools/route-test` DIRECT/RELAY benchmark. GitHub Actions publishes Linux relay
and Windows executable artifacts after successful tests. UDP/TCP and ticket
rejection smoke use real authenticated traffic; local CI is not Internet proof.

## Run the local demo

Install Go 1.25+ and run:

```sh
go run ./tools/demo/cmd/demo
```

Open **http://127.0.0.1:8099**. The dark frontend has Home, Applications, Routes,
Statistics, Diagnostics and Settings. It uses real loopback UDP traffic through
three real ticket-authenticated QUIC relays. DIRECT has deliberately configured
45 ms delay and one-in-five response loss; relay paths have 5/10/15 ms delays.
Measured RTT/loss may vary with scheduling. No random or invented telemetry is
rendered. Wait for measurements and hysteresis, add a local application profile,
then use **Stop relay** to exercise actual relay failure and new-flow failover.
Optimize toggles the lab application's routing; it does not capture host programs.
The lab process identity is explicit test metadata, not Windows discovery.

`go run ./tools/lab` runs the original minimal UDP/TCP end-to-end lab.

## Run the backend development environment

Install Docker with Compose v2:

```sh
docker compose up --build -d
docker compose --profile test run --rm dev-client
```

This starts PostgreSQL 17, the control plane, ticket-required relay and isolated
UDP/TCP echo targets. The test client really registers, verifies signed config,
obtains a device-bound ticket and sends both protocols through the relay.
Health: http://127.0.0.1:8080/healthz; metrics: ports 8080 and 9090.
Secrets are generated once into scoped Docker volumes, never source files.
Keep these private: dev-client is a bootstrap/admin integration tool, not the
normal least-privilege client. HTTP, private-CIDR egress and the test CA are
**development only**. `docker compose down` preserves state; deleting volumes
also deletes identities, keys and the database.

[Deployment](DEPLOYMENT.md) includes optional Prometheus/Grafana and cloud build
CA configuration. [Control plane](CONTROL_PLANE.md) and [relay](RELAY.md) describe
standalone processes, endpoints and credentials.

## Build and test

```sh
go build ./...
go test -race -timeout 120s ./...
go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...
```

Go race checks need a C compiler. Set `NEEXITLAG_TEST_DATABASE_URL` to a disposable
PostgreSQL database to run API integration tests; otherwise that test explicitly
skips. CI provisions PostgreSQL. Short fuzz commands:

```sh
go test ./core/protocol -run '^$' -fuzz '^FuzzDecode$' -fuzztime 5s -parallel 2
go test ./core/protocol -run '^$' -fuzz '^FuzzReadOpen$' -fuzztime 5s -parallel 2
go test ./core/security -run '^$' -fuzz '^FuzzTicket$' -fuzztime 10000x -fuzzminimizetime 1s -parallel 1
go test ./core/security -run '^$' -fuzz '^FuzzConfig$' -fuzztime 10000x -fuzzminimizetime 1s -parallel 1
go test ./core/ipc -run '^$' -fuzz '^FuzzFrame$' -fuzztime 5s -parallel 2
```

Browser smoke: install Python Playwright and Chromium, run the demo, then
`python3 tests/frontend/demo_smoke.py` (override `CHROMIUM_PATH` if needed).
It changes the local demo profile/settings and stops one real relay.

## Windows desktop and service

Use Windows 10/11 x64, Go 1.25+, .NET 8 SDK and trusted, signed WinDivert v2 binaries:

```powershell
go build -o bin/neexitlag-service.exe ./services/client-service/cmd/service
dotnet build apps/desktop/Neexitlag.Desktop.csproj
```

Provision a device with `go run ./tools/enroll -control-plane https://YOUR-CP
-state DEVICE-DIR -admin-token-file BOOTSTRAP-TOKEN-FILE` (one command line),
then run [install.ps1](tools/windows/install.ps1) elevated with the service binary,
owner SID, device/trust/driver directories, trusted DLL/SYS SHA256 pins,
HTTPS control plane URL and literal IP:port UDP echo target. `-TCPProbeTarget`
optionally enables the TCP diagnostic. The script installs SCM service `neexitlag`
and protects ProgramData credentials; run the desktop without elevation.
[Windows acceptance](docs/BLOCKERS.md) gives the remaining runtime test matrix.

WPF can also bind to the portable lab on Windows. Save the demo HTML's
`const token='...'` value into a private local token file, then start:

```powershell
dotnet run --project apps/desktop -- --demo-url http://127.0.0.1:8099 --demo-token-file TOKEN-FILE
```

That token belongs only to the current local demo instance. Native desktop bindings
were exercised on Linux by `tests/desktop-bindings` using the same C# client source;
this is not evidence that a native WPF window or named-pipe driver session ran.

## Network quality v1

AUTO compares 60-second rolling route history, with sustained loss and burst loss
penalties, confirmed improvements and flow pinning. Native route/statistics views
show available quality metrics; diagnostics distinguish observable local drops
from unanswered probes. See [policy, metrics and validation](docs/NETWORK_QUALITY_V1.md)
and [destination quality / disabled FEC design](docs/design/DESTINATION_QUALITY_AND_FEC.md).

## Known limitations

Probe quality describes the configured echo destination, not every game server.
The capture backend supports unfragmented IPv4 UDP/TCP. IPv4 loopback and unknown
identities remain DIRECT. IPv6 relay transport is unsupported: a network guard
blocks attributable new relay-app IPv6 flows; existing/unattributed
sockets fail open, so this is not a leak-proof privacy guarantee. Established TCP
connections cannot migrate across egress IPs; failures close them and new
connections follow policy. Profile publisher/signature requirements safely do
not match until a verified publisher is supplied. WFP, signed application updates, updater/channel rollout,
production identity enrollment, DDoS infrastructure, driver signing and Internet
performance validation are future work. Update channel is a persisted preference;
there is no implemented updater. The MTU diagnostic verifies a 1100-byte UDP echo,
not discovered Internet path MTU. Config expiry and relay session leases bound
operation during a control-plane outage; reconnection requires valid tickets.

User-facing brand: **VELIQ**. Compatibility and safe Windows upgrade: [VELIQ branding](docs/VELIQ_BRANDING.md). Repository/module, SCM, IPC and persisted paths intentionally retain their established identifiers.
