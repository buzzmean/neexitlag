# Chrome AUTO by origin — PR #8

No extension is required. This mode is a **Chrome-only loopback HTTP CONNECT/PAC
transport**. It does not change Windows/WinHTTP/WinINET proxy settings, routes,
DNS servers, certificates, credentials or the VPS. The existing authenticated QUIC
relay pool and same-session lease renewal are used without protocol changes.

## Why the previous implementation was wrong

Chrome AUTO could select an app-wide optimizer winner or a configured YouTube
baseline before reading SNI/Host. Now every native Chrome AUTO TCP connection
waits for domain classification and its decision is pinned. A YouTube baseline
cannot authorize DIRECT+ for OpenAI or ordinary sites. Standalone ChatGPT/Codex,
Edge and Discord keep their existing application policies.

SNI alone cannot guarantee origin routing: ECH hides the inner name, and HTTP/2
can send another origin over a connection whose first SNI was different. Request
metadata alone cannot change that connection's egress safely. Native opaque TCP
uses authenticated required relay (missing SNI is not an open failure); opaque Chrome UDP uses authenticated required relay,
never a TCP proof or DNS/IP guess. For complete origin attribution, activate the
Chrome CONNECT/PAC mode below.

## Activate without an extension

1. Update/build/install the Windows service and GUI using the existing upgrade
   procedure (preserve ProgramData identity/credentials).
2. Select Chrome by its full executable path in **Applications**.
3. Click **Chrome: AUTO по сайтам** (the ordinary AUTO button uses the same path
   for Chrome). Save unfinished work and accept a graceful Chrome restart.
4. The GUI configures only the current user's **Google Chrome ProxySettings**
   policy and launches Chrome with the same persistent default user-data directory,
   restoring the previous session. It never kills processes or copies/deletes the
   Chrome profile. Windows proxy remains untouched. Foreign/organization proxy
   policy is refused rather than overwritten.
5. **Отключить PAC Chrome** now switches Chrome AUTO back to DIRECT and retains
   an owned mandatory protected-domain guard; OpenAI never gets DIRECT fallback.
   It does not blindly remove security policy. Foreign policies are refused. The default standard Google Chrome profile is supported by the GUI;
   custom/portable user-data directories need an explicit launch configuration.

An inline `data:` PAC has no network download failure that silently falls back to
DIRECT. The owned Chrome policy also sets `ProxyPacMandatory=true`, so a PAC
evaluation error is fail-closed. Online policy returns only the local proxy for every origin, including DIRECT.
DIRECT upstreams never need VPS availability. There is no online `; DIRECT`
alternative: Chrome otherwise caches a failed proxy and keeps bypassing the
healthy transport after restart. The owned **separate recovery policy** returns
DIRECT for ordinary origins while retaining strict proxy guards for OpenAI,
YouTube and explicit non-DIRECT domains. Non-domain rules retain strict mode. Each canonical
hostname uses a distinct SHA256-derived `*.localhost` proxy name. Chrome resolves
these names internally to loopback, without system DNS overrides. Distinct proxy
identities partition Chrome connection pools, including HTTP/2 coalescing. The
local service reads the CONNECT hostname **before any origin connection or client
TLS bytes** leave the machine. TLS stays end-to-end between Chrome and the site.

The endpoint admits only current, selected Chrome process identities belonging to
the installation's owner SID, using the existing bounded TCP owner table and
process discovery. It is not an unauthenticated public/local-user relay endpoint.

The non-secret loopback port is persisted in the existing service state directory
and restored **before** automatic capture startup. Verified profile IDs are
persisted with this non-secret port in the ACL-protected state directory. It stays stable across service
restarts. While the GUI is running, unavailable IPC switches the owned policy to recovery.
It restores the online policy (and the prior Optimize state) when the listener
returns; Chrome remains running. A live capture with a failed listener degrades
only Chrome to guarded recovery and shows an activation error. Without the GUI,
a stopped proxy is unavailable until service restart or explicit GUI recovery;
no automatic browser bypass is advertised. Unknown origins selected by reachability learning are
best-effort AUTO, not an offline security rule; use an explicit required rule for
a domain that must remain protected when the service is absent. Optimization disable/re-enable closes
only helper-owned connections so future connections converge without changing an
established socket's egress. Initial PAC activation may restart Chrome, as allowed
for this mode. Other applications and their sockets are untouched.

## Automatic decisions

* `chatgpt.com`, `openai.com`, `oaistatic.com`, `oaiusercontent.com`, `chat.com` and
  their subdomains: **RELAY_REQUIRED**, default VPS-01. Domain suffix boundaries
  are exact; `openai.com.evil.example` does not match. These rules cannot be
  downgraded by a learned DIRECT result or another domain's custom rule.
* YouTube/Googlevideo and the existing curated YouTube asset domains: the existing
  operator-confirmed **DIRECT+** strategy, bound to the same network and revision.
  It is reused only for these origins, not all Chrome. The native WinDivert engine
  handles an exact delegated upstream tuple under its original Chrome identity.
* Unknown ordinary origins start with **DIRECT**. A bounded synthetic `HEAD /`
  request checks a real HTTP status and normal TLS certificate validation without
  cookies, authentication, path/query or page-content inspection. On a network/TLS
  transport failure it tries a domain-only clone of the existing verified local
  strategy, then authenticated relay. No local candidate is invented as a winner;
  missing candidates are skipped. DIRECT+ success also requires actual native
  modified-packet/byte evidence. A TCP ACK alone is never success.
* An HTTP 4xx/5xx application/WAF/auth response is not PASS and does not trigger
  repeated login/egress retries. Certificate trust failures are left to Chrome;
  no certificate exception or MITM is added in production.

Trial evidence includes route, HTTP status/failure class, elapsed time and actual
transform proof. Planning/timeouts remain visible as failures, not missing rows.

Trials are scoped by application, exact domain, port/scheme, network and policy
revision. They use at most three real attempts, 1.5 seconds each, four probe
workers, a 32-origin pending bound and a 512-entry expiring cache. Successful
origin evidence lasts 15 minutes; terminal failures cool down for one minute.
Other domains/networks do not inherit the result. Only new upstream connections
use the selected route; application bytes are never replayed across routes.
The transport admits at most 256 client connections, retains at most 512 metadata
records and bounds headers to 16 KiB. IPC views report truncation explicitly.

## Evidence and limits

The Applications screen lists **domain, route, reason and origin-probe result**.
A connected tunnel is `TCP_ESTABLISHED_HTTP_RESULT_UNKNOWN`, not page PASS.
DIRECT+ stays "not confirmed" until real WinDivert transform evidence exists.
Failed required opens show BLOCK. No request bodies, cookies, page contents or
TLS keys enter diagnostics.

* A successful origin HEAD probe does **not** prove that a logged-in web app,
  streamed video or every API path succeeds. The product does not claim this.
* Chrome HTTPS and WebSocket TLS remain encrypted; HTTP/2 runs inside CONNECT.
  This HTTP CONNECT mode uses TCP, so browser HTTP/3 falls back to TCP. The relay's
  QUIC/mTLS `/3` transport and other applications' UDP are unchanged. This is not
  a claim of native per-origin HTTP/3 support.
* Chrome CONNECT carries the hostname even with ECH/DoH. Origin resolution uses
  the service's existing system resolver; DNS servers are not changed. DNS
  privacy and bypass of ISP DNS poisoning are not promised. Resolver failures are
  explicit. Private/literal origins do not trigger automatic Internet probes.
* IPv4 is preferred for native DIRECT+. IPv6 DIRECT/relay can use existing TCP
  transports when available; native IPv6 DIRECT+ is unsupported. No transparent
  IPv6 runtime support is claimed from cross-builds.
* Chrome launched without the PAC/policy is not equivalent to this mode. Native
  SNI guards cannot prove HTTP/2 inner-origin behavior; the GUI's domain table is
  evidence only for actual CONNECT traffic. Managed/foreign proxy conflicts are
  refused. Keep the service installed while its Chrome policy is active.

## Automated acceptance (no long manual voice/browser test)

`TestChromePACRealHTTPSOriginIsolation` runs an actual isolated Chrome/Chromium,
HTTP/2 TLS target and authenticated relay on loopback. Nine origins (the seven public hostnames plus ordinary/blocked test domains) share one IP
and certificate. A target nonce returned to browser JavaScript is mandatory;
route intent, TCP ACK or JSON alone cannot PASS. An artificially impaired unknown
origin rejects local source traffic but accepts the relay's distinct source IP.
Public OpenAI/YouTube and the production VPS are never contacted by the fixture.

`TestChromeSitesNativeAcceptance` additionally runs the actual Windows service
controller, TCP owner/SID checks, signed WinDivert and the existing disorder
strategy. It requires positive native modified packet/byte counters for YouTube,
real HTTP(S) responses, authenticated protected traffic, ordinary DIRECT and
complete evidence. It disables/re-enables optimization, stops the full service
controller/WinDivert backend, verifies ordinary HTTPS during outage with OpenAI
fail-closed, restarts the controller on the saved port, and verifies new responses
in the **same Chrome process**. The target also verifies real ECH/GREASE ClientHellos (bounded ephemeral parsing,
no payload logs). Windows Chrome runs with its normal sandbox; only
the root Linux container needs `--no-sandbox`. Every YouTube CONNECT must have its own
positive native transform evidence. Missing browser/driver/proof cannot be
reported as PASS.
The loopback fixture uses 1000-byte socket writes to model Internet MTU; its
jumbo-MSS packets otherwise exceed the existing engine’s 1500-byte packet bound.
These are real packets, not synthetic counters; TLS features remain enabled.
Loopback packet observation is enabled only through a test-binary constructor,
not a production flag. The Chrome acceptance uses a temporary browser directory
and a pinned ephemeral test certificate. It also exercises actual current-user
Chrome PAC policy without the proxy launch flags; the disposable CI user’s value
is removed afterward, and any foreign policy is refused. Production never enables CDP or ignores
certificate errors.

```powershell
$env:NEEXITLAG_TEST_DRIVER_DIR=(Resolve-Path .e2e/windivert).Path
go test -v -timeout 120s ./platform/windows -run '^TestChromeSitesNativeAcceptance$'
```

The native acceptance is enabled automatically after the existing driver download
in Verify CI. Domain-policy, cache/cooldown/stale-evidence, PAC isolation,
certificate validation, stream preservation and shutdown regressions run in the
normal Go suite. CI loopback PASS is not public Internet performance evidence.


## Regression ff9585a: confirmed cause and activation transaction

The real Windows report showed `browser.status.listening=false`, no VK CONNECT
record, and native `BLOCK / POLICY_BLOCK / ECH_OUTER_SNI_ONLY`. AUTO had been
published before CONNECT/PAC activation. `ResolveDeferredDecision` then treated
missing usable SNI as a failed protected open and blocked even an explicit
opaque-name authenticated relay route. Modern Chrome can send ECH GREASE too;
we do not guess whether its inner name is readable or disable ECH globally.
A second code defect routed Chrome DNS/53 through the opaque relay policy,
including queries to the LAN resolver. DNS bootstrap now preserves its existing
DIRECT path and does not grant any TCP/443 or QUIC bypass.

Activation now proceeds in this order:

1. Save/quit Chrome normally and install a mandatory fail-closed OpenAI recovery
   guard; no profile or global Optimize change yet.
2. `browser.prepare` binds the bounded loopback listener and returns a 128-bit,
   45-second, one-use readiness URL. Bind failure leaves routing unchanged.
3. Install the owned current-user online PAC. An isolated, empty headless Chrome
   profile requests the readiness URL through the **actual policy**, without proxy
   launch flags; real user tabs are not restored before AUTO is ready. TCP owner,
   creation identity, executable profile and owner SID are checked before proof.
4. `browser.confirm` verifies that exact nonce. IPC reachability, a listening port
   or the GUI's own TCP connection cannot certify Chrome readiness.
5. Publish Chrome AUTO, enable capture, then launch/restore the real user profile. A new AUTO profile and Optimize with
   an unactivated Chrome AUTO are rejected before capture state changes. Optimize
   also verifies the actual owner-SID HKCU mandatory online PAC, not merely a
   listening port or persisted readiness ID. The GUI
   checks command results and cannot label a rejected command as success.
6. Failure installs the guarded recovery PAC and restores a safe Chrome profile
   (legacy/broken AUTO becomes DIRECT). Domain rules are preserved in
   `browser_recovery_rules` and restored on AUTO reactivation; the guarded recovery
   profile must not itself defer ordinary ECH back into opaque BLOCK. Confirmed
   activation proof is revoked when leaving AUTO. Other applications/Optimize are not
   globally disabled. OpenAI stays blocked/required through the guard, never
   DIRECT. An incomplete rollback gets an explicit error, not PASS.

Native TCP resolves the current exact tuple owner before trusting cached FLOW
associations. A still-live old process is not proof of port ownership: ephemeral
port reuse must not turn service DIRECT/CONNECT sockets into recursive Chrome
interception. Only exact owned DIRECT+ delegations override the service exclusion.

`browser.status.failures` now distinguishes TCP owner lookup, token access, SID
mismatch, profile mismatch, malformed/timeout proxy headers, DNS planning,
upstream failure classes and capacity. No credentials, arbitrary OS error strings,
HTTP headers or payload are logged. Failed ordinary synthetic HEAD probes are
not proof that the browser must use an unavailable relay; only confirmed origin
results authorize automatic escalation. Explicit required rules are untouched.

## Windows update after ff9585a

Close Chrome and the old GUI normally first. In an elevated PowerShell at the
repository root (no uninstall, enrollment or VPS action):

```powershell
git switch codex/direct-plus-flowseal
git pull --ff-only origin codex/direct-plus-flowseal
if ($LASTEXITCODE) { throw 'git pull failed' }
git rev-parse HEAD
.\tools\windows\build-e2e.ps1
.\tools\windows\build-desktop.ps1
Stop-Service neexitlag
(Get-Service neexitlag).WaitForStatus('Stopped', [TimeSpan]::FromSeconds(30))
Copy-Item .\bin\windows-e2e\neexitlag-service.exe "$env:ProgramData\neexitlag\neexitlag-service.exe" -Force
Start-Service neexitlag
(Get-Service neexitlag).WaitForStatus('Running', [TimeSpan]::FromSeconds(30))
```

Run **the newly built** `bin/release/NeExitLag.exe` as the installation's normal
Windows user. Select Chrome → **Chrome: AUTO по сайтам** and allow its graceful
restart/readiness tab. Merely enabling Optimize is not Chrome activation. On a
legacy installation without a verified PAC state, automatic capture intentionally
refuses to arm and reports the missing Chrome activation. Existing identity,
certificates, profiles, local strategy and all VPS files remain intact.

The controlled-origin CI tests are transport/routing evidence, not public
VK/YouTube/OpenAI availability or logged-in app/video playback acceptance. The
reported public-site Windows failures are not declared PASS from these fixtures.

After GUI activation, a short public-site check (one isolated Chrome process,
normal sandbox/TLS, **actual HKCU policy**, no proxy command-line override) is:

```powershell
.\tools\windows\acceptance-chrome-sites.ps1
# Optional, elevated: two 30-second rounds with a local SCM service restart,
# preserving the exact diagnostic Chrome process; never contacts SSH/VPS admin.
.\tools\windows\acceptance-chrome-sites.ps1 -RestartService
```

It opens VK, ChatGPT, Twitch, Yandex Music, langame, Telegram Web and YouTube
simultaneously, checks completed real main-document HTTP responses plus API/script
network failures, observes HTTP/2, requires attributed route/transform evidence,
and drains bounded capture change cursors. WAF/403, empty/error pages, instrumentation
loss, missing evidence or queue drops are FAIL. It logs no page text, cookies,
headers or TLS keys (only DOM length/readiness and public origin metadata).
This helper has not been run against the user's public Internet connection in CI;
controlled-origin acceptance is separate. A PASS is not logged-in functionality,
Discord voice or full video playback proof.
