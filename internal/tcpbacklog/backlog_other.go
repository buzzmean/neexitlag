//go:build !windows

package tcpbacklog

import "net"

// Unix Go listeners already use the OS maximum (e.g. Linux somaxconn). This
// helper addresses the different Winsock meaning of SOMAXCONN only.
func configure(*net.TCPListener) error { return nil }
