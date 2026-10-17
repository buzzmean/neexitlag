//go:build windows

package windowsplatform

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// dialBoundIPv4 registers the exact source tuple before Go emits a SYN. Never
// bind inside Dialer.Control: Go binds LocalAddr after Control and ConnectEx
// binds again when LocalAddr is nil, so a manual bind fails with WSAEINVAL.
// Reserve an ephemeral port, release it, then let Go perform the ONLY bind on
// the outbound socket. An intervening port collision has no SYN/payload and is
// retried at most twice with a new port; other errors are never retried. Go owns
// the socket, IOCP, cancellation, half-close and all resource cleanup normally.
func dialBoundIPv4(parent context.Context, source netip.Addr, dest netip.AddrPort, bound func(netip.AddrPort) error) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if !source.Is4() || !dest.Addr().Is4() {
		return nil, errors.New("bound IPv4 socket required")
	}
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		reservation, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: source.AsSlice()})
		if err != nil {
			return nil, err
		}
		local := reservation.Addr().(*net.TCPAddr).AddrPort()
		if err = reservation.Close(); err != nil {
			return nil, err
		}
		dialer := net.Dialer{LocalAddr: net.TCPAddrFromAddrPort(local), ControlContext: func(ctx context.Context, network, address string, raw syscall.RawConn) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return bound(local)
		}}
		conn, err := dialer.DialContext(ctx, "tcp4", dest.String())
		if err == nil {
			return conn, nil
		}
		if !errors.Is(err, windows.WSAEADDRINUSE) {
			return nil, err
		}
	}
	return nil, errors.New("bounded upstream source-port collision")
}
