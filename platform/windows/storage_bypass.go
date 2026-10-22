package windowsplatform

import "fmt"

func IsPrivateISCSI(t Tuple) bool {
	return t.Protocol == "tcp" && (t.Destination.Port() == 3260 && t.Destination.Addr().IsPrivate() || t.Source.Port() == 3260 && t.Source.Addr().IsPrivate())
}
func privateIPv4Filter(field string) string {
	return fmt.Sprintf("((%s >= 10.0.0.0 and %s <= 10.255.255.255) or (%s >= 172.16.0.0 and %s <= 172.31.255.255) or (%s >= 192.168.0.0 and %s <= 192.168.255.255))", field, field, field, field, field, field)
}

// Network-layer exclusion: diskless storage never enters the redirect queue.
// Match the private STORAGE peer, not the always-private client source address.
func nativeTCPNetworkFilter(port uint16, loopbackFixture bool) string {
	base := fmt.Sprintf("ip and tcp and !impostor and ((outbound and !loopback) or (inbound and (!loopback or tcp.DstPort == %d)))", port)
	if loopbackFixture {
		base = "ip and tcp and !impostor"
	}
	return base + " and not ((tcp.DstPort == 3260 and " + privateIPv4Filter("ip.DstAddr") + ") or (tcp.SrcPort == 3260 and " + privateIPv4Filter("ip.SrcAddr") + "))"
}
