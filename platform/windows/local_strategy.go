package windowsplatform

import (
	"encoding/binary"
	"github.com/buzzmean/neexitlag/core/localstrategy"
)

// SendLocal shares the existing NETWORK send path, with no additional handles.
// Any local engine failure falls back to the original packet. TCP sequence space
// is unchanged, so a partially sent split followed by the original is harmless
// TCP retransmission, not replay on a new connection.
func SendLocal(d Decision, state *localstrategy.State, id ProcessIdentity, tuple Tuple, p []byte, send func([]byte) error, evidence *EvidenceLedger) error {
	if d.Local == nil || !d.Route.Local() || id.PID == 0 || id.CreationTime == 0 || id.Path == "" {
		err := send(p)
		if err == nil {
			evidence.RecordBytes(id, tuple, "direct", "", uint64(len(p)))
		}
		return err
	}
	r := d.Local.Process(p, state, localstrategy.Metadata{StrategyID: d.LocalStrategyID, Application: id.Path, Protocol: tuple.Protocol, Destination: tuple.Destination})
	if r.Modified {
		for _, q := range r.Packets {
			if err := send(q); err != nil {
				d.Local.Fail(state)
				r = localstrategy.Result{Failed: true}
				break
			}
		}
		if r.Modified {
			d.Local.Commit(r, len(p))
			evidence.RecordLocal(id, tuple, r.Strategy, r.Domain, uint64(len(p)))
			return nil
		}
	}
	err := send(p)
	if err == nil {
		d.Local.Commit(r, len(p))
		if r.Strategy != "" && state.Applied() {
			evidence.RecordLocal(id, tuple, r.Strategy, r.Domain, uint64(len(p)))
		} else {
			evidence.RecordBytes(id, tuple, "direct", "", uint64(len(p)))
		}
	}
	return err
}

// The engine computes UDP checksum before splitting. WinDivert's helper cannot
// compute a transport checksum from a fragment; preserve that valid checksum and
// mark it complete rather than allowing checksum-offload metadata from the input
// to describe a different (fragmented) packet. Only used on validated UDP path.
func LocalPacketAddress(p []byte, a [80]byte) [80]byte {
	if len(p) >= 20 && p[0] == 0x45 && p[9] == 17 && binary.BigEndian.Uint16(p[6:])&0x3fff != 0 {
		a[10] |= 0xa0
	}
	return a
}
