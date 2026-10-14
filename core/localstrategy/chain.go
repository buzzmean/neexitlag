package localstrategy

import (
	"encoding/binary"
	"errors"
	"sort"
	"strings"
)

// Actions are a finite packet plan, not a scripting language. At most one fake,
// one segmentation and one reversal; at most five emitted packets per input.
type Action struct {
	Kind      string   `json:"kind"`
	Positions []uint16 `json:"positions,omitempty"`
	Anchor    string   `json:"anchor,omitempty"` // empty=payload, sni=complete SNI offset
	TTL       uint8    `json:"ttl,omitempty"`
}

func validateChain(p Profile) error {
	if p.Protocol != "tcp" || len(p.Actions) < 1 || len(p.Actions) > 3 || p.SplitPosition != 0 {
		return errors.New("invalid TCP chain")
	}
	split, fake, reverse := false, false, false
	for i, a := range p.Actions {
		switch a.Kind {
		case "fake-tls":
			if fake || i != 0 || a.TTL < 2 || a.TTL > 16 || len(a.Positions) > 0 || a.Anchor != "" {
				return errors.New("fake TLS requires explicit TTL 2..16 and first position")
			}
			fake = true
		case "tcp-multisplit":
			if split || reverse || a.TTL != 0 || len(a.Positions) < 1 || len(a.Positions) > 3 || (a.Anchor != "" && a.Anchor != "sni") {
				return errors.New("invalid bounded split")
			}
			last := uint16(0)
			for _, v := range a.Positions {
				if v == 0 || v > 512 || v <= last {
					return errors.New("split positions must increase within 1..512")
				}
				last = v
			}
			split = true
		case "tcp-disorder":
			if !split || reverse || i != len(p.Actions)-1 || len(a.Positions) > 0 || a.TTL != 0 || a.Anchor != "" {
				return errors.New("disorder requires prior split")
			}
			reverse = true
		default:
			return errors.New("unknown chain action")
		}
	}
	return nil
}
func ChainLabel(actions []Action) string {
	parts := []string{}
	for _, a := range actions {
		parts = append(parts, a.Kind)
	}
	return strings.Join(parts, " + ")
}

// tlsSNI locates only a COMPLETE single-record ClientHello. No QUIC decryption,
// ECH inner-name attribution, stream reassembly, retained payload or guessing.
func tlsSNI(p []byte) (int, int, bool) {
	if len(p) < 44 || p[0] != 22 || p[1] != 3 || p[2] < 1 || p[2] > 3 || p[5] != 1 || p[9] != 3 || p[10] < 1 || p[10] > 3 || p[43] > 32 {
		return 0, 0, false
	}
	end := 5 + int(binary.BigEndian.Uint16(p[3:]))
	if end > len(p) || end < 44 {
		return 0, 0, false
	}
	handshake := int(p[6])<<16 | int(p[7])<<8 | int(p[8])
	if handshake+9 != end {
		return 0, 0, false
	}
	i := 43
	take := func(n int) bool {
		if n < 0 || i+n > end {
			return false
		}
		i += n
		return true
	}
	if !take(1) || !take(int(p[43])) || !take(2) {
		return 0, 0, false
	}
	n := int(binary.BigEndian.Uint16(p[i-2:]))
	if n < 2 || n%2 != 0 || !take(n) || !take(1) {
		return 0, 0, false
	}
	n = int(p[i-1])
	if n < 1 || !take(n) || !take(2) {
		return 0, 0, false
	}
	n = int(binary.BigEndian.Uint16(p[i-2:]))
	if n != end-i {
		return 0, 0, false
	}
	start, size := 0, 0
	for i+4 <= end {
		k := binary.BigEndian.Uint16(p[i:])
		n = int(binary.BigEndian.Uint16(p[i+2:]))
		i += 4
		if i+n > end {
			return 0, 0, false
		}
		// Match only the observable outer SNI. ECH inner names are never read
		// or inferred; a public_name outside the explicit set stays DIRECT.
		if k == 0 {
			if start != 0 || n < 5 || int(binary.BigEndian.Uint16(p[i:]))+2 != n || p[i+2] != 0 {
				return 0, 0, false
			}
			size = int(binary.BigEndian.Uint16(p[i+3:]))
			if size+5 != n {
				return 0, 0, false
			}
			start = i + 5
		}
		i += n
	}
	if i != end || start == 0 {
		return 0, 0, false
	}
	if _, err := domain(string(p[start : start+size])); err != nil {
		return 0, 0, false
	}
	return start, size, true
}

// Own deterministic TLS1.2 ClientHello, not copied upstream data. Deliberately
// innocuous example.org SNI; fake is always entirely OLD sequence space.
func fakeHello() []byte {
	host := []byte("www.example.org")
	body := append([]byte{3, 3}, make([]byte, 32)...)
	body = append(body, 0, 0, 2, 0, 0x2f, 1, 0)
	ext := []byte{0, 0, 0, byte(5 + len(host)), 0, byte(3 + len(host)), 0, 0, byte(len(host))}
	ext = append(ext, host...)
	body = append(body, byte(len(ext)>>8), byte(len(ext)))
	body = append(body, ext...)
	handshake := append([]byte{1, byte(len(body) >> 16), byte(len(body) >> 8), byte(len(body))}, body...)
	return append([]byte{22, 3, 1, byte(len(handshake) >> 8), byte(len(handshake))}, handshake...)
}
func oldSequenceFake(p []byte, st *State, ttl uint8) []byte {
	payload, ok := packetPayload(p, "tcp")
	if !ok || p[33]&0x10 == 0 || !st.synSeen || !st.handshakeConfirmed || st.inspected != 1 || st.fakeSent || binary.BigEndian.Uint32(p[24:]) != st.synSequence+1 || binary.BigEndian.Uint32(p[28:]) != st.serverSequence+1 || p[8] <= ttl {
		return nil
	}
	hello := payload
	if st.logicalHello != nil {
		hello = st.logicalHello
	}
	if _, _, ok = tlsSNI(hello); !ok {
		return nil
	}
	data := fakeHello()
	q := make([]byte, 40+len(data))
	copy(q, p[:40])
	copy(q[40:], data)
	q[32] = 0x50
	q[33] = 0x18 // no SYN/FIN/RST/URG or options/negotiation changes
	q[8] = ttl
	binary.BigEndian.PutUint16(q[2:], uint16(len(q)))
	binary.BigEndian.PutUint16(q[4:], binary.BigEndian.Uint16(p[4:])+1)
	// Inclusive fake end = SYN ISN; RCV.NXT is at least SYN ISN+1.
	// Modulo arithmetic is TCP serial arithmetic, including ISN wraparound.
	binary.BigEndian.PutUint32(q[24:], st.synSequence+1-uint32(len(data)))
	checksumTransport(q)
	return q
}
func chainPackets(p []byte, st *State, c *compiled) Result {
	payload, ok := packetPayload(p, "tcp")
	if !ok || p[33]&0x10 == 0 || len(payload) == 0 || len(p) > 1500 {
		return Result{}
	}
	var real [][]byte
	var fake []byte
	var ttl uint8
	for _, a := range c.Actions {
		switch a.Kind {
		case "fake-tls":
			fake = oldSequenceFake(p, st, a.TTL)
			if fake == nil {
				return Result{}
			} // uncertain handshake => entire plan fail-open
			ttl = a.TTL
		case "tcp-multisplit":
			offset := 0
			if a.Anchor == "sni" {
				var ok bool
				offset, _, ok = tlsSNI(payload)
				if !ok {
					return Result{}
				}
			}
			positions := []int{0}
			for _, v := range a.Positions {
				pos := offset + int(v)
				if pos <= 0 || pos >= len(payload) {
					return Result{}
				}
				positions = append(positions, pos)
			}
			positions = append(positions, len(payload))
			sort.Ints(positions)
			h := len(p) - len(payload)
			real = nil
			for i := 0; i+1 < len(positions); i++ {
				lo, hi := positions[i], positions[i+1]
				q := make([]byte, h+hi-lo)
				copy(q, p[:h])
				copy(q[h:], payload[lo:hi])
				if i+2 < len(positions) {
					q[33] &^= 8
				}
				if i > 0 {
					q[33] &^= 0x80
				}
				binary.BigEndian.PutUint32(q[24:], binary.BigEndian.Uint32(p[24:])+uint32(lo))
				binary.BigEndian.PutUint16(q[2:], uint16(len(q)))
				checksumTransport(q)
				real = append(real, q)
			}
		case "tcp-disorder":
			for i, j := 0, len(real)-1; i < j; i, j = i+1, j-1 {
				real[i], real[j] = real[j], real[i]
			}
		}
	}
	if len(real) == 0 {
		q := append([]byte(nil), p...)
		checksumTransport(q)
		real = [][]byte{q}
	}
	if fake != nil {
		st.fakeSent = true
		real = append([][]byte{fake}, real...)
	}
	return Result{Packets: real, Modified: len(real) > 1, FakePackets: boolCount(fake != nil), FakeTTL: ttl, Chain: ChainLabel(c.Actions), Preset: c.Preset}
}
func boolCount(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}

func observeSYN(p []byte, st *State) {
	if len(p) < 40 || p[0] != 0x45 || p[9] != 6 || int(binary.BigEndian.Uint16(p[2:])) != len(p) || binary.BigEndian.Uint16(p[6:])&0x3fff != 0 || p[33]&0x3f != 2 {
		return
	}
	h := 20 + int(p[32]>>4)*4
	if h != len(p) || h < 40 {
		return
	} // SYN payload/Fast Open is not eligible for fake
	seq := binary.BigEndian.Uint32(p[24:])
	if st.synSeen && st.synSequence != seq {
		st.Disable()
		return
	}
	st.synSeen = true
	st.synSequence = seq
}

// Called only for an exact reverse-tuple inbound SYN/ACK observed by the existing
// TCP capture handle. It grants no route decision and stores no packet payload.
func (s *State) ConfirmHandshake(clientACK, serverSequence uint32) {
	if s.synSeen && !s.failed && clientACK == s.synSequence+1 {
		s.handshakeConfirmed = true
		s.serverSequence = serverSequence
	}
}
