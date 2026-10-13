# Windows networking boundary

`WindowsCaptureBackend` separates service/routing from the native capture backend.
WinDivert v2 integration joins FLOW identity with NETWORK IPv4 UDP packets, pins
first-packet decisions, checks PID creation time and endpoint deletion, repairs
checksums and injects replies into the original application tuple. Unknown/self,
TCP/IPv6/fragments and unselected traffic remain DIRECT. Queues, flow identity
state and relay forwarding concurrency are bounded.

Absolute-path loader requires trusted DLL/SYS SHA256 pins, validates signed SYS
Authenticode and constrains library search. External signed binaries are not
committed. Missing dependencies fail before capture. WFP is a distinct unavailable
boundary awaiting a signed kernel driver, preserving ADR-001.

Windows cross-build and cross-compiled discovery tests passed. This is implemented
source, not verified live interception. See [acceptance blockers](../../docs/BLOCKERS.md).

Real E2E tools, owner-table fallback (including dual-stack socket ownership),
process cache/cleanup and counter evidence are described in
[Windows real E2E](../../docs/WINDOWS_REAL_E2E.md). Official unsigned DLL is protected
by pinned SHA256; signature validation applies to the signed kernel SYS.
