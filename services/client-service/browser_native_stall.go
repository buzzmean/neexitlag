package clientservice

import (
	"net"
	"sync"
	"time"
)

// Unlike the legacy PAC ClientHello retry, native forwarding never replays or
// moves established traffic. An initial silent DIRECT connection is closed;
// only a subsequent browser connection can use the failed-domain cache.
type nativeFirstReplyConn struct {
	net.Conn
	mu                     sync.Mutex
	timer                  *time.Timer
	sent, answered, closed bool
	timeout                time.Duration
	stalled                func()
}

func (c *nativeFirstReplyConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.mu.Lock()
	if n > 0 {
		c.sent = true
		if !c.answered && !c.closed && c.timer == nil {
			c.timer = time.AfterFunc(c.timeout, c.fail)
		}
	}
	c.mu.Unlock()
	return n, err
}
func (c *nativeFirstReplyConn) fail() {
	c.mu.Lock()
	if c.closed || c.answered || !c.sent {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()
	c.Conn.Close()
	if c.stalled != nil {
		c.stalled()
	}
}
func (c *nativeFirstReplyConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.mu.Lock()
		c.answered = true
		if c.timer != nil {
			c.timer.Stop()
		}
		c.mu.Unlock()
	} else if err != nil {
		c.fail()
	}
	return n, err
}
func (c *nativeFirstReplyConn) Close() error {
	c.mu.Lock()
	c.closed = true
	if c.timer != nil {
		c.timer.Stop()
	}
	c.mu.Unlock()
	return c.Conn.Close()
}
func (c *nativeFirstReplyConn) CloseWrite() error {
	if w, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return w.CloseWrite()
	}
	return nil
}
