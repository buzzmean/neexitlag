//go:build windows

package windowsplatform

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/buzzmean/neexitlag/core/gameudp"
	"github.com/buzzmean/neexitlag/core/routing"
)

// Native PID/birth lookup + the real capture dispatch. Only final WinDivert
// writes are replaced; this is not actual-driver or game/anti-cheat acceptance.
func TestNativeGameUDPFirstFlightAndUnselectedProcess(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestHotAttachIdentityChild$", "-test.v")
	cmd.Env = append(os.Environ(), "NEEXITLAG_HOT_IDENTITY_CHILD=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stdin.Close(); cmd.Wait() })
	ready := make(chan struct{})
	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			if s.Text() == "IDENTITY_READY" {
				close(ready)
				return
			}
		}
	}()
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("identity helper")
	}
	id, err := NewDiscovery().Lookup(uint32(cmd.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	for _, selected := range []bool{false, true} {
		b, _ := captureFixture(t)
		b.gameGuard.ready = true
		e := gameudp.New()
		e.Configure(id.Path, true)
		e.SetReason("")
		p, tuple := gamePacket()
		b.associations.Put(Association{Identity: id, Tuple: tuple})
		fake, original := 0, 0
		b.packetWriter = func(q []byte, _ [80]byte) error {
			if bytes.Equal(p, q) {
				original++
			} else {
				fake++
			}
			return nil
		}
		classify := func(_ ProcessIdentity, tuple Tuple) Decision {
			d := Decision{Flow: routing.Flow{PID: id.PID, ProcessPath: id.Path, Protocol: "udp", Source: tuple.Source, Destination: tuple.Destination}, Revision: "game", Network: strings.Repeat("a", 64)}
			if selected {
				d.GameUDP = e
			}
			return d
		}
		for i := 0; i < 4; i++ {
			b.processUDP(nil, p, [80]byte{}, classify)
		}
		want := 0
		if selected {
			want = 24
		}
		if fake != want || original != 4 {
			t.Fatal(fake, original, selected)
		}
		raw, err := json.Marshal(b.evidence.Snapshot(id.PID))
		if err != nil || len(raw) == 0 {
			t.Fatal(err)
		}
	}
}
