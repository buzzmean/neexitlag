package clientservice

// Called with browser.mu held. Native readiness must not inherit PAC readiness
// or fabricate a listener/nonce. Only NATIVE_TCP counters establish native ingress.
func (c *Controller) nativeBrowserReadiness(edge bool) string {
	for _, p := range c.Profiles.List() {
		if !nativeBrowserProfile(p) || edgeProfile(p) != edge {
			continue
		}
		c.mu.Lock()
		enabled := c.enabled && c.settings.Automatic
		c.mu.Unlock()
		if !enabled {
			return "NATIVE_DISABLED"
		}
		if !c.nativeBackendReady() {
			return "NATIVE_CAPTURE_NOT_READY"
		}
		if !c.relayConnected() {
			return "NATIVE_RELAY_PENDING"
		}
		c.browser.native.mu.Lock()
		m := c.browser.native.profiles[p.ID].NativeOnly
		c.browser.native.mu.Unlock()
		if m.Exchanges > 0 {
			return "NATIVE_ONLY_OBSERVED"
		}
		if m.Opened > 0 {
			return "NATIVE_OPENED_NO_EXCHANGE"
		}
		return "NATIVE_READY_NEW_CONNECTIONS"
	}
	return ""
}
