//go:build windows

package windowsplatform

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// dialBoundIPv4 registers the exact bound tuple BEFORE ConnectEx emits a SYN.
// net.Dialer.Control cannot do this: Go binds again after Control (and again in
// ConnectEx setup when LocalAddr is nil), so a manually bound socket fails with
// WSAEINVAL. Use an overlapped connect, then hand a duplicate to Go's net poller.
// No payload is sent here; cancellation waits for the OS operation to complete
// before releasing its OVERLAPPED/event or socket.
func dialBoundIPv4(parent context.Context, source netip.Addr, dest netip.AddrPort, bound func(netip.AddrPort) error) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !source.Is4() || !dest.Addr().Is4() {
		return nil, errors.New("bound IPv4 socket required")
	}
	fd, err := windows.WSASocket(windows.AF_INET, windows.SOCK_STREAM, windows.IPPROTO_TCP, nil, 0, windows.WSA_FLAG_OVERLAPPED|windows.WSA_FLAG_NO_HANDLE_INHERIT)
	if err != nil {
		return nil, os.NewSyscallError("WSASocket", err)
	}
	owned := true
	defer func() {
		if owned {
			windows.Closesocket(fd)
		}
	}()
	if err = windows.Bind(fd, &windows.SockaddrInet4{Addr: source.As4()}); err != nil {
		return nil, os.NewSyscallError("bound_bind", err)
	}
	sa, err := windows.Getsockname(fd)
	if err != nil {
		return nil, os.NewSyscallError("bound_getsockname", err)
	}
	local, ok := sa.(*windows.SockaddrInet4)
	if !ok {
		return nil, errors.New("invalid bound IPv4 socket")
	}
	if err = bound(netip.AddrPortFrom(source, uint16(local.Port))); err != nil {
		return nil, err
	}
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return nil, os.NewSyscallError("connect_event", err)
	}
	defer windows.CloseHandle(event)
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	ov := &windows.Overlapped{HEvent: event}
	var sent, flags uint32
	err = windows.ConnectEx(fd, &windows.SockaddrInet4{Addr: dest.Addr().As4(), Port: int(dest.Port())}, nil, 0, &sent, ov)
	if err != nil && !errors.Is(err, windows.ERROR_IO_PENDING) {
		return nil, os.NewSyscallError("bound_connectex", err)
	}
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { windows.CancelIoEx(fd, ov); close(done) })
	err = windows.WSAGetOverlappedResult(fd, ov, &sent, true, &flags)
	if !stop() {
		<-done
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, os.NewSyscallError("bound_connect_result", err)
	}
	if err = windows.Setsockopt(fd, windows.SOL_SOCKET, windows.SO_UPDATE_CONNECT_CONTEXT, nil, 0); err != nil {
		return nil, os.NewSyscallError("bound_connect_context", err)
	}
	file := os.NewFile(uintptr(fd), "neexitlag-bound-tcp")
	if file == nil {
		return nil, errors.New("bound socket file unavailable")
	}
	owned = false
	defer file.Close()
	conn, err := net.FileConn(file)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
