// Package routing owns decisions and flow metadata independently of packet capture.
package routing

import (
	"context"
	"net/netip"
	"time"
)

type Route struct{ RelayID string }

func (r Route) Direct() bool { return r.RelayID == "" }

type Flow struct {
	Protocol                   string
	Source, Destination        netip.AddrPort
	PID                        uint32
	ProcessPath, ApplicationID string
	Route                      Route
	CreatedAt, LastActivity    time.Time
}

// Backend must identify the process before passing its flow to the routing core.
// Stop must remove interception and restore DIRECT networking.
type Backend interface {
	Run(context.Context, func(Flow) Route) error
	Stop() error
}
type Profile struct {
	ID, Name            string
	Executables         []string
	Policy              string
	ProtocolHints       []string
	PortHints           []uint16
	TransportPreference string
}
