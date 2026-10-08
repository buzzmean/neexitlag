# Windows networking boundary

`WindowsCaptureBackend` separates service/routing from the native capture backend.
WinDivert v2 integration joins FLOW identity with NETWORK IPv4 UDP packets, pins
first-packet decisions, checks PID creation time and endpoint deletion, repairs
checksums and injects replies into the original application tuple. Unknown/self,
TCP/IPv6/fragments and unselected traffic remain DIRECT. Queues, flow identity
state and relay forwarding concurrency are bounded.

Absolute-path loader requires trusted DLL/SYS SHA256 pins, validates DLL
Authenticode and constrains library search. External signed binaries are not
committed. Missing dependencies fail before capture. WFP is a distinct unavailable
boundary awaiting a signed kernel driver, preserving ADR-001.

Windows cross-build and cross-compiled discovery tests passed. This is implemented
source, not verified live interception. See [acceptance blockers](../../docs/BLOCKERS.md).
