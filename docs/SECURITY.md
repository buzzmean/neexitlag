# Threat model and current controls
| Threat | Current control / outstanding work |
|---|---|
| Stolen token | No bearer tokens in Phase 1; short-lived tickets planned |
| Stolen device | mTLS key identity; revoke CA-issued leaf via future registry |
| Fake relay / MITM | Trusted CA, hostname/IP verification, TLS 1.3 |
| Malicious relay | Can read forwarded payload at egress; applications need end-to-end crypto |
| Malicious client | Verified device cert, bounded sessions/flows/windows, destination policy |
| Replay / session hijack | QUIC packet protection, no 0-RTT; future tickets need key binding |
| Config tampering | Signed config + Last Known Good planned, not implemented |
| Update compromise | Signed updates and driver release pipeline planned |
| DDoS / amplification | QUIC handshake validation, authenticated DATAGRAM, packet token bucket; infrastructure protection still needed |

Logs must exclude payload, credentials, chat messages and URLs. Relay startup
logs only listener address or errors. Metrics contain counters, no identities.
Server only accepts literal destination IPs, avoiding server DNS rebinding.
Use explicit destination allowlists and host firewall egress limits in production.
CA keys are never stored in the repository. labcert creates short-lived test keys
in memory; they are not a provisioning system. Read docs/STATUS.md for verified tests.
