# Real Windows / public-Internet acceptance

**Status: READY FOR VPS. Actual Windows capture and Internet acceptance have not
been performed.** Cloud has no Windows runtime, public VPS address, SSH binding or
permitted VPS TCP destination. A successful build or local lab is not acceptance.
Phase 0–6 checkpoint fa7cdc1 is preserved; pre-work checkpoint is 80eef79.

This stage changes Windows IPv4 UDP capture/discovery, diagnostics, ordinary test
applications and deployment/acceptance tooling. GUI, scoring, the relay wire
protocol and the existing demo are retained.

## Windows preparation

Use a dedicated Windows 10/11 x64 test PC, elevated PowerShell 5.1+ or PowerShell7,
Go1.25+, Git and Windows OpenSSH client. SSH needs an existing key/agent and a
verified known_hosts entry. The VPS login must be root or have noninteractive
sudo. Do not paste private keys or tokens into chat. Synchronize both host clocks.

Clone a fresh checkout so existing working files are preserved:

```powershell
git clone --branch codex/windows-real-e2e https://github.com/buzzmean/neexitlag.git neexitlag-e2e
cd neexitlag-e2e
Set-ExecutionPolicy -Scope Process RemoteSigned
```

Set `$VPS_PUBLIC_IP`, `$SSH_USER`, `$SSH_PORT` to the operator's actual values.
There is no example IP default; private, loopback and documentation addresses are
rejected by the deployment/acceptance commands.

```powershell
.\tools\windows\setup-e2e.ps1 -VpsPublicIP $VPS_PUBLIC_IP -SSHUser $SSH_USER -SSHPort $SSH_PORT
.\tools\windows\acceptance.ps1 -VpsPublicIP $VPS_PUBLIC_IP -SSHUser $SSH_USER -SSHPort $SSH_PORT
```

Setup builds the Windows service, selected/control executables and Linux backend
binaries, fetches official pinned WinDivert, creates scoped private credentials,
uploads via verified SSH, installs VPS units, enrolls the device using private-CA
HTTPS and installs/recovery-configures the Windows service. It refuses to replace
an existing Windows service or unrelated VPS identity. A same-VPS private bundle
can be reused after an ingress/enrollment failure. Keep `.e2e` private; it contains
operator CA/admin credentials as well as client state. No key material is printed.
The service receives only its device identity/certificate, trust, probe token and
IPC credentials, never the CA signing key or bootstrap admin token.

Open provider/host firewall ingress before enrollment: TCP8443 (HTTPS), UDP4433
(QUIC), UDP40000 (authenticated echo). Preserve SSH access. Setup does not enable
UFW, rewrite existing firewall rules, or change Windows default routes. For an
already active UFW installation, the corresponding rules are:

```sh
sudo ufw allow 8443/tcp
sudo ufw allow 4433/udp
sudo ufw allow 40000/udp
```

Provider security-group rules are separate. VPS minimum: Ubuntu24.04 LTS or Debian
with systemd, x86_64, public IPv4, approximately1GB RAM and a few GB free disk.
The public echo endpoint must also be reachable from the VPS itself. Acceptance
measures actual VPS echo egress rather than guessing its source IP, accommodating
provider NAT when public-IP hairpin works. Unsupported hairpin produces a clear
failure; it is not treated as route proof.

## Evidence and acceptance results

`test-game.exe` and `control-client.exe` are two copies of one ordinary UDP test
program with different absolute executable paths and real PIDs. Neither embeds a
relay client or supplies process metadata to the service. The kernel capture
backend identifies socket owners, executable paths and creation times. Both send
128-byte authenticated sequence/nonce probes to the same public target.

The echo target returns its actual observed source address/port, authenticated by
HMAC-SHA256. Replies are no larger than requests; unknown/bad/reply packets are
rejected and replies are limited to100/s (burst200). This is an authenticated
acceptance target, not an anonymous public echo/reflector or relay proxy.

The service's `capture.diagnostics` IPC method reports real counters:
`captured_packets`, `relayed_packets`, `direct_packets`, `dropped_packets`,
`reinject_packets`, `loop_prevention_packets`, `relay_reply_packets`, plus PID,
absolute executable, creation time, source/destination, selected route and relay.
Global reinjection counts include restored relay replies. Loop-prevention counts
are observed service-origin packets explicitly bypassed by PID, not fabricated
recapture statistics. Payloads/tokens are not logged. Flow evidence is bounded,
expires after5minutes, and is frame-size limited; truncated reports are flagged.

Acceptance prints PASS/FAIL for:

- A: selected exe in MANUAL RELAY, verified capture counters/replies and VPS echo
  egress. This does not depend on GUI status or manufactured RTT.
- B: unselected control exe DIRECT, with a different echo source and zero relay
  counters for its real PID/creation time.
- C: A and B process lifetimes overlap.
- F: service QUIC transport is recognized by its actual PID and bypass counters;
  its packets never enter application relay routing.
- D: AUTO must first actually choose the healthy relay from real measurements;
  stopping the remote relay must yield native DIRECT packets and both egress
  sources within the same application's report. The script waits through normal
  route lifetime/confirmation. If AUTO prefers DIRECT, D is FAIL/NOT EXERCISED;
  it never forces a score or injects synthetic impairment to claim a transition.
- Short control-plane outage: existing authenticated session and valid LKG survive,
  while the control application remains DIRECT. Expired-cache behavior is a
  separate unit-test scope, not an Internet runtime claim.
- E: an active control stream survives graceful service stop and forced process
  termination; SCM recovery restarts capture.
- Driver dependency error: a temporary wrong DLL hash rejects capture, while the
  control stream remains DIRECT. This is validation-failure acceptance, not a
  simulated kernel-driver crash.
- Default routes remain identical.

Reports and per-client metadata are in `.e2e/reports`. Exit0/overall PASS requires
all stages, including D. Any prerequisite failure creates FAIL/NOT RUN entries.
The script restores settings/manifest, removes only its temporary custom profiles,
restarts remote services it stopped, and records restoration failures. It stops
and kills the test service deliberately; use a dedicated test PC.

## Manual runs and operations

```powershell
.\tools\windows\build-e2e.ps1
.\tools\windows\fetch-windivert.ps1
.\tools\windows\start.ps1
.\tools\windows\stop.ps1
.\tools\windows\uninstall.ps1
```

Uninstall removes the neexitlag SCM entry and preserves credentials/driver state
for explicit rollback. It does not remove a shared WinDivert driver.

Normal application probing, with service profile deciding its route:

```powershell
.\bin\windows-e2e\test-game.exe -target "${VPS_PUBLIC_IP}:40000" -token-file .\.e2e\bundle\windows\echo.token -count 100 -interval 100ms -report .\.e2e\game.json
.\bin\windows-e2e\control-client.exe -target "${VPS_PUBLIC_IP}:40000" -token-file .\.e2e\bundle\windows\echo.token -count 100 -interval 100ms -report .\.e2e\control.json
```

The program displays destination, sequence, sent/received, monotonic RTT, jitter,
loss and observed echo sources. Comparing the two reports compares actual DIRECT
and RELAY paths. The app never secretly selects a route through its own tunnel.

VPS operations:

```sh
sudo systemctl status neexitlag-control-plane neexitlag-relay neexitlag-echo
sudo systemctl restart neexitlag-relay
sudo systemctl stop neexitlag-relay
sudo systemctl start neexitlag-relay
sudo journalctl -u neexitlag-relay --since '10 minutes ago'
curl --fail http://127.0.0.1:9090/healthz
curl --fail http://127.0.0.1:9090/metrics
curl --fail http://127.0.0.1:9092/metrics
```

TLS control-plane health uses `/etc/neexitlag/control/ca.pem` with curl --cacert.
Metrics bind loopback, not public ingress. CP, relay and echo use separate Linux
users and private file scopes. PostgreSQL listens locally; CA/admin/signing keys
are unavailable to relay/echo users. Relay permits only the test VPS destination
CIDR, enforces mTLS + device-bound tickets and persists its replay cache.
Certificate expiry30days is intentional; production renewal is not implemented.

## Windows boundary and remaining verification

Native code captures outbound unfragmented IPv4 UDP, preserving application
socket tuple in replies. FLOW events provide identity; a unique owner from
GetExtendedUdpTable resolves first-packet races. Both IPv4 and IPv6 ownership
tables are queried because an IPv4 packet can originate from a dual-stack socket.
This does **not** implement IPv6 interception. Ambiguous/unknown/exited identities
remain DIRECT. PID cache checks creation time on hits and has cleanup. Socket/
flow association and capture pins are bounded and idle-cleaned.

The service explicitly bypasses its own PID; reinjected impostor packets are
excluded from the capture filter. Failed relay setup/send/receive switches to
DIRECT by reinjecting original application packets and draining pending payloads.
Existing TCP is never magically migrated; transparent TCP/IPv6 is not implemented.
An oversized first datagram stays DIRECT. Oversized datagrams on an already
pinned RELAY flow are dropped under the retained1100-byte transport budget;
fragments bypass interception. These limits must be assessed before real games. Service/driver abrupt failure behavior still
needs actual Windows A–F reports. Linux cannot establish kernel reinjection safety.

After A–F pass, add an absolute `cs2.exe` custom profile through authenticated IPC
with mode AUTO. The routing engine needs no rewrite and no game IP is hardcoded.
The test VPS relay allowlist currently permits only its echo host; enabling other
public destinations is a separate deliberate operator policy change, retaining
mTLS/tickets. No real CS2 session or Internet improvement has been verified.

[Cloud verification and its limits](WINDOWS_E2E_CLOUD_VERIFICATION.md).
