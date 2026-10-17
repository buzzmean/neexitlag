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
is therefore fail-closed; opaque Chrome UDP uses authenticated required relay,
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
5. Use **Отключить PAC Chrome** and restart Chrome before uninstalling the service.
   Only the policy value owned by NeExitLag is removed. Foreign changes are not
   deleted. The default standard Google Chrome profile is supported by the GUI;
   custom/portable user-data directories need an explicit launch configuration.

An inline `data:` PAC has no network download failure that silently falls back to
DIRECT. The owned Chrome policy also sets `ProxyPacMandatory=true`, so a PAC
evaluation error is fail-closed. It returns only a local proxy, without a DIRECT
fallback. Each canonical
hostname uses a distinct SHA256-derived `*.localhost` proxy name. Chrome resolves
these names internally to loopback, without system DNS overrides. Distinct proxy
identities partition Chrome connection pools, including HTTP/2 coalescing. The
local service reads the CONNECT hostname **before any origin connection or client
TLS bytes** leave the machine. TLS stays end-to-end between Chrome and the site.

The endpoint admits only current, selected Chrome process identities belonging to
the installation's owner SID, using the existing bounded TCP owner table and
process discovery. It is not an unauthenticated public/local-user relay endpoint.

The non-secret loopback port is persisted in the existing service state directory
and restored after normal optimization startup. It stays stable across service
restarts. A stopped/unavailable service makes the PAC connection fail; protected
traffic never acquires a DIRECT fallback. Optimization disable/re-enable closes
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
HTTP/2 TLS target and authenticated relay on loopback. Four origins share one IP
and certificate. A target nonce returned to browser JavaScript is mandatory;
route intent, TCP ACK or JSON alone cannot PASS. An artificially impaired unknown
origin rejects local source traffic but accepts the relay's distinct source IP.
Public OpenAI/YouTube and the production VPS are never contacted by the fixture.

`TestChromeSitesNativeAcceptance` additionally runs the actual Windows service
controller, TCP owner/SID checks, signed WinDivert and the existing disorder
strategy. It requires positive native modified packet/byte counters for YouTube,
real HTTP(S) responses, authenticated protected traffic, ordinary DIRECT and
complete evidence. It also disables/re-enables optimization and verifies new
responses in the same Chrome process. Every YouTube CONNECT must have its own
positive native transform evidence. Missing browser/driver/proof cannot be
reported as PASS.
Loopback packet observation is enabled only through a test-binary constructor,
not a production flag. The Chrome acceptance uses a temporary browser directory
and a pinned ephemeral test certificate; production never enables CDP or ignores
certificate errors.

```powershell
$env:NEEXITLAG_TEST_DRIVER_DIR=(Resolve-Path .e2e/windivert).Path
go test -v -timeout 120s ./platform/windows -run '^TestChromeSitesNativeAcceptance$'
```

The native acceptance is enabled automatically after the existing driver download
in Verify CI. Domain-policy, cache/cooldown/stale-evidence, PAC isolation,
certificate validation, stream preservation and shutdown regressions run in the
normal Go suite. CI loopback PASS is not public Internet performance evidence.
