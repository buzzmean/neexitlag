# Windows backend boundary
Phase 3 is not implemented. The routing.Backend contract isolates interception.
Production target: WFP ALE process classification plus signed callout driver;
UDP datagram redirection and TCP connect redirection need separate implementations.
MVP candidate: WinDivert FLOW/SOCKET metadata joined to NETWORK packets, with
PID reuse protection, reinjection loop prevention, checksum repair, IPv6 and
fail-open shutdown verified on Windows. Never change the default route.
No stub backend claims to intercept Windows traffic.
