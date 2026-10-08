# Implementation and verification status

Phase 0–2 commits remain ancestors of this work: b74718d, 14a93d1,
7967f8352e810d39dea6c793b0bf8cfa20d027f8. Pre-change checkpoint: aa33a03.

| Area | Implemented | Verification / remaining work |
|---|---|---|
| Original data plane | mTLS QUIC UDP/TCP, pinned DIRECT/RELAY, limits and fallback | Existing Go integration suite and original lab retained |
| Phase 4 | Scoring, PRIMARY/SECONDARY/TERTIARY, hysteresis, bounded history, adaptive real UDP probes | Deterministic selection tests; real echo/failover tests |
| Phase 5 | PostgreSQL migration/entities/APIs, registration, signed ticket, heartbeat, signed configs/LKG, profiles | Actual PostgreSQL integration and Docker UDP/TCP ticket flow |
| Phase 3 | Process path/PID/creation identity, FLOW association, WinDivert loader, IPv4 UDP reinjection, SCM, authenticated named pipes | Windows cross-build and cross-compiled Windows tests; native runtime unverified |
| WFP | Separate explicit unavailable backend boundary | Kernel driver not implemented |
| Phase 6 | Dark six-page WPF desktop, .exe picker, real service bindings; portable browser lab | WPF build, shared C# client runtime test, browser E2E passed; WPF window unverified |
| Observability | Relay/CP Prometheus endpoints, probe/health demo metrics, optional Grafana provisioning | Local metrics and real heartbeat/lease tests |

No benchmark or Internet latency improvement claim. No external Windows runtime
was available. Transparent TCP/IPv6 interception and service-crash/driver behavior
remain acceptance blockers. GitHub Windows CI is configured without a capture
driver; its result must be checked on the PR, not inferred from local cross-build.
