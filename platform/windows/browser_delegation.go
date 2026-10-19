package windowsplatform

import "time"

// Full close of a planned CONNECT frontend owns both halves of this native
// mapping. A TCP FIN alone is insufficient: a silent origin can retain the
// Engine reservation forever while the local proxy waits for its response.
// Only the exact delegation generation may release it. CloseWrite is unchanged.
func (s *tcpState) closePlannedDelegation(bound Tuple, owner *tcpEntry) {
	s.mu.Lock()
	e := s.entries[bound]
	if e == nil || e.delegationOwner != owner || e.decision.OpenDeferred == nil || e.mode == "closed" || e.mode == "hot_attach" {
		s.mu.Unlock()
		return
	}
	up, local, cancel := e.upstream, e.local, e.openCancel
	e.upstream, e.local = nil, nil
	e.mode, e.last, e.normalTeardown = "closed", time.Now(), true
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	// Release the Engine key synchronously, before the caller may retry the
	// same origin through relay. No wait for a remote FIN or pipe goroutine.
	if up != nil {
		_ = up.Close()
	}
	if local != nil {
		_ = local.Close()
	}
}
