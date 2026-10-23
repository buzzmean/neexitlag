package clientservice

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/buzzmean/neexitlag/core/routing"
)

// A signed, well-known P0 primary node has its OWN ticket audience/replay cache.
// VPS-01 is only the local policy/provider name; its tickets are NEVER reused for
// VPS-PRIMARY. Existing opened sockets retain their original provider.
const primaryRelayID = "VPS-PRIMARY"

type primaryRelayPool interface {
	routing.Relay
	Healthy() bool
}
type primaryUsage struct {
	ID                    string `json:"authenticated_relay_id"`
	Address               string `json:"configured_address"`
	Endpoint              string `json:"observed_tcp_session_endpoint"`
	Opens, Active, TX, RX uint64
}
type relayPrimaryRuntime struct {
	mu               sync.Mutex
	rows             map[string]primaryUsage
	primary, standby string
	last, reason     string
}

func newRelayPrimary() *relayPrimaryRuntime {
	return &relayPrimaryRuntime{rows: map[string]primaryUsage{}}
}
func (r *relayPrimaryRuntime) snapshot() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	rows := []primaryUsage{}
	for _, v := range r.rows {
		rows = append(rows, v)
	}
	return map[string]any{"policy_relay_id": "VPS-01", "configured_primary": r.primary, "configured_standby": r.standby, "last_opened_authenticated_relay": r.last, "reason": r.reason, "connections": rows, "existing_tcp": "PINNED; no replay/migration; configured primary alone is not traffic proof", "ticket_audiences": "unique server IDs; fresh ticket per session; no shared logical audience on two servers"}
}

type primaryRelayProvider struct {
	primary, standby               primaryRelayPool
	primaryAddress, standbyAddress string
	runtime                        *relayPrimaryRuntime
}

func (p *primaryRelayProvider) order() (primaryRelayPool, string, string, primaryRelayPool, string, string) {
	if !p.primary.Healthy() && p.standby != nil && p.standby.Healthy() {
		return p.standby, "VPS-01", p.standbyAddress, p.primary, primaryRelayID, p.primaryAddress
	}
	return p.primary, primaryRelayID, p.primaryAddress, p.standby, "VPS-01", p.standbyAddress
}
func (p *primaryRelayProvider) opened(id, address, endpoint, reason string) func() {
	r := p.runtime
	r.mu.Lock()
	v := r.rows[id]
	v.ID = id
	v.Address = address
	if endpoint != "UDP_ENDPOINT_NOT_EXPOSED_BY_POOL" {
		v.Endpoint = endpoint
	}
	v.Opens++
	v.Active++
	r.rows[id] = v
	r.last = id
	r.reason = reason
	r.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			r.mu.Lock()
			v := r.rows[id]
			if v.Active > 0 {
				v.Active--
			}
			r.rows[id] = v
			r.mu.Unlock()
		})
	}
}
func (p *primaryRelayProvider) OpenTCP(ctx context.Context, dst netip.AddrPort) (net.Conn, error) {
	first, id, address, second, other, otherAddress := p.order()
	attempt, stop := context.WithTimeout(ctx, 1500*time.Millisecond)
	conn, err := first.OpenTCP(attempt, dst)
	stop()
	reason := "SIGNED_PRIMARY_HEALTHY"
	if id != "VPS-PRIMARY" {
		reason = "PRIMARY_UNAVAILABLE_STANDBY"
	}
	if err != nil && ctx.Err() == nil && second != nil {
		conn, err = second.OpenTCP(ctx, dst)
		id, address = other, otherAddress
		reason = "NEW_OPEN_FAILED_OTHER_AUTHENTICATED_RELAY"
	}
	if err != nil {
		return nil, err
	}
	closeCount := p.opened(id, address, conn.RemoteAddr().String(), reason)
	return &primaryMeter{Conn: conn, runtime: p.runtime, id: id, release: closeCount}, nil
}
func (p *primaryRelayProvider) OpenDatagram(ctx context.Context, dst netip.AddrPort) (routing.Datagram, error) {
	first, id, address, second, other, otherAddress := p.order()
	attempt, stop := context.WithTimeout(ctx, 1500*time.Millisecond)
	flow, err := first.OpenDatagram(attempt, dst)
	stop()
	reason := "SIGNED_PRIMARY_HEALTHY"
	if id != primaryRelayID {
		reason = "PRIMARY_UNAVAILABLE_STANDBY"
	}
	if err != nil && ctx.Err() == nil && second != nil {
		flow, err = second.OpenDatagram(ctx, dst)
		id, address = other, otherAddress
		reason = "NEW_OPEN_FAILED_OTHER_AUTHENTICATED_RELAY"
	}
	if err != nil {
		return nil, err
	}
	return &primaryDatagram{Datagram: flow, runtime: p.runtime, id: id, release: p.opened(id, address, "UDP_ENDPOINT_NOT_EXPOSED_BY_POOL", reason)}, nil
}

type primaryMeter struct {
	net.Conn
	runtime *relayPrimaryRuntime
	id      string
	release func()
}

func (m *primaryMeter) Read(b []byte) (int, error) {
	n, e := m.Conn.Read(b)
	m.runtime.add(m.id, 0, uint64(n))
	return n, e
}
func (m *primaryMeter) Write(b []byte) (int, error) {
	n, e := m.Conn.Write(b)
	m.runtime.add(m.id, uint64(n), 0)
	return n, e
}
func (m *primaryMeter) Close() error { e := m.Conn.Close(); m.release(); return e }
func (m *primaryMeter) CloseWrite() error {
	if v, ok := m.Conn.(interface{ CloseWrite() error }); ok {
		return v.CloseWrite()
	}
	return nil
}
func (r *relayPrimaryRuntime) add(id string, tx, rx uint64) {
	r.mu.Lock()
	v := r.rows[id]
	v.TX += tx
	v.RX += rx
	r.rows[id] = v
	r.mu.Unlock()
}

type primaryDatagram struct {
	routing.Datagram
	runtime *relayPrimaryRuntime
	id      string
	release func()
}

func (d *primaryDatagram) Send(b []byte) error {
	e := d.Datagram.Send(b)
	if e == nil {
		d.runtime.add(d.id, uint64(len(b)), 0)
	}
	return e
}
func (d *primaryDatagram) Receive(ctx context.Context) ([]byte, error) {
	b, e := d.Datagram.Receive(ctx)
	d.runtime.add(d.id, 0, uint64(len(b)))
	return b, e
}
func (d *primaryDatagram) Close() error { e := d.Datagram.Close(); d.release(); return e }

// These are routing providers only. They cannot return DIRECT, even when both
// authenticated relays fail. Explicit flowpolicy fail-closed is unchanged.
