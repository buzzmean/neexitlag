# Benchmarks
No throughput or latency improvement claims yet. Planned: 64/512/1100-byte UDP,
TCP streams, packets/sec, Gbps, CPU, memory per flow and transport overhead.
Protocol payload limit is 1100 bytes, so 1200-byte forwarding is rejected until
PMTU-aware support is implemented. Run protocol benchmarks with go test -bench . ./core/protocol.
