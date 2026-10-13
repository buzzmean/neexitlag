# Windows E2E preparation: cloud verification

Linux x64, Go1.25.1, PowerShell7.4.6, actual PostgreSQL17.
No real Windows runtime or operator VPS was available. Status: **READY FOR VPS**.

| Check | Result / actual scope |
|---|---|
| go build ./... | PASS |
| go test -race -count=1 -timeout 120s ./... | PASS with PostgreSQL URL supplied |
| go vet ./... | PASS |
| Windows amd64 CGO0 build of all Go packages | PASS |
| Windows vet of platform/service packages | PASS |
| Windows discovery/UDP owner test executable | Compiled; not executed on Linux |
| build-e2e.ps1 | PASS, produced five Windows executables and four Linux binaries |
| PowerShell syntax | All published setup/acceptance/service/test scripts PASS |
| Actual PowerShell IPC client | Linux .NET fixture: framed round-trip, bad ID and bounded timeout PASS |
| VPS shell installer | bash -n PASS; not executed against a remote host |
| systemd units | systemd-analyze verify PASS in isolated dependency/executable fixtures |
| Official WinDivert download | Archive and DLL/SYS SHA256 verified; PE signature-table structure inspected |
| Original tools/lab | Actual local UDP/TCP E2E PASS |
| Real Windows A–F | NOT RUN |
| Public-Internet relay/echo | NOT RUN |
| Real CS2 | NOT RUN |

The new integration test actually runs PostgreSQL APIs, private-CA HTTPS device
registration, signed config/ticket issuance, mTLS QUIC relay and authenticated
nonce echo over both DIRECT and relay transports. It then validates persisted LKG
with the HTTP server stopped. It deliberately asserts that this portable transport
alone does not advertise Windows capture as connected.

Other tests cover PID reuse/cache expiry, wildcard/shared-port ownership,
IPv4-mapped/dual-stack ownership parsing, service-PID exclusion and counter-ledger
cleanup, echo authentication/reply rejection/observed source, actual ordinary UDP
client RTT/loss and temporary profile removal without deleting default profiles.
Windows-specific tests use real current-process discovery and actual sockets;
only their compilation was checked here. CI can run those without a driver, but
that still cannot substitute for administrator-run public-Internet acceptance.

A new IPC fixture caught an AsyncWaitHandle ownership error in the PowerShell
client; removing disposal of framework-owned cached handles resolved it, and all
three scenarios subsequently passed. The new service-LKG integration fixture
initially checked an uninitialized helper cache; it now loads and checks the
actual signed cache persisted by the service. These were resolved before the
final regression. No saved report is represented as a Windows runtime PASS.
