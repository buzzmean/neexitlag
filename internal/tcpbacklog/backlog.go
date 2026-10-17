// Package tcpbacklog makes Windows pending-handshake capacity explicit. Winsock
// SOMAXCONN means an implementation-defined "reasonable maximum", not the
// numeric maximum. This must not undercut bounded application admission.
package tcpbacklog

import (
	"fmt"
	"net"
)

const Limit = 512 // >= 384 pending setups; fixed, never user-controlled/unbounded

func Configure(l *net.TCPListener) error {
	if err := configure(l); err != nil {
		return fmt.Errorf("configure bounded TCP listen backlog: %w", err)
	}
	return nil
}
