package windowsplatform

import (
	"encoding/binary"
	"errors"
	"net/netip"
)

func udpOwner(b []byte, source netip.AddrPort) (uint32, error) {
	if len(b) < 4 {
		return 0, errors.New("truncated UDP table")
	}
	n := uint64(binary.LittleEndian.Uint32(b))
	if n > (uint64(len(b))-4)/12 {
		return 0, errors.New("invalid UDP table size")
	}
	var owner uint32
	for i := uint64(0); i < n; i++ {
		r := b[4+i*12 : 4+(i+1)*12]
		var ip [4]byte
		copy(ip[:], r[:4])
		a := netip.AddrFrom4(ip)
		if binary.BigEndian.Uint16(r[4:6]) != source.Port() || (!a.IsUnspecified() && a != source.Addr()) {
			continue
		}
		pid := binary.LittleEndian.Uint32(r[8:12])
		if pid == 0 || owner != 0 && owner != pid {
			return 0, errors.New("ambiguous UDP owner")
		}
		owner = pid
	}
	if owner == 0 {
		return 0, errors.New("UDP owner not found")
	}
	return owner, nil
}
