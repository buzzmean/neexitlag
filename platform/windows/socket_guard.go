package windowsplatform

import (
	"encoding/binary"
	"errors"
	"net/netip"
)

// DecodeSocketEvent reads native SOCKET_BIND/CONNECT events. A
// connect gives us identity before a SYN, including processes with dual-stack
// sockets. UDP IPv6 binds must fail rather than silently bypass RELAY selection.
func DecodeSocketEvent(a [80]byte) (uint32, Tuple, error) {
	if a[8] != 3 || (a[9] != 3 && a[9] != 4) {
		return 0, Tuple{}, errors.New("not socket bind/connect")
	}
	data := a[16:]
	protocol := "tcp"
	if data[56] == 17 {
		protocol = "udp"
	} else if data[56] != 6 {
		return 0, Tuple{}, errors.New("unsupported protocol")
	}
	decode := func(b []byte) netip.Addr {
		if a[10]&16 == 0 {
			var ip [4]byte
			binary.BigEndian.PutUint32(ip[:], binary.LittleEndian.Uint32(b[12:16]))
			return netip.AddrFrom4(ip)
		}
		var ip [16]byte
		for i := 0; i < 4; i++ {
			binary.BigEndian.PutUint32(ip[i*4:], binary.LittleEndian.Uint32(b[(3-i)*4:]))
		}
		return netip.AddrFrom16(ip)
	}
	return binary.LittleEndian.Uint32(data[16:20]), Tuple{protocol, netip.AddrPortFrom(decode(data[20:36]), binary.LittleEndian.Uint16(data[52:54])), netip.AddrPortFrom(decode(data[36:52]), binary.LittleEndian.Uint16(data[54:56]))}, nil
}
func BlockUnsupportedIPv6(id ProcessIdentity, t Tuple, event byte, self uint32, classify Classifier) bool {
	// TCP bind is admitted; the destination is known at connect. UDP bind is
	// denied for selected relay apps because unconnected SendTo has no connect.
	if t.Protocol == "tcp" && event == 3 {
		return false
	}
	d, bypass := ClassifyCapture(id, t, self, classify)
	return !bypass && (d.Block || !d.Route.Direct())
}
