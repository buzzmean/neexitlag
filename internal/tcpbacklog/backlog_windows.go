//go:build windows

package tcpbacklog

import (
	"net"
	"syscall"
)

func configure(l *net.TCPListener) error {
	raw, err := l.SyscallConn()
	if err != nil {
		return err
	}
	var listenErr error
	if err = raw.Control(func(fd uintptr) {
		// SOMAXCONN_HINT(N) = -N. Unlike SOMAXCONN, this requests the explicit
		// supported Winsock backlog (Windows bounds hints to 200..65535).
		// Re-listening adjusts this socket before its accept loop is started.
		listenErr = syscall.Listen(syscall.Handle(fd), -Limit)
	}); err != nil {
		return err
	}
	return listenErr
}
