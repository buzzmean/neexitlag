package flowpolicy

import "strings"

// BrowserRelayGroup is a bounded family hint, not authorization for DIRECT and
// not evidence that the family is blocked. Never match generic Google domains.
func BrowserRelayGroup(host string) string {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, suffix := range []string{"youtube.com", "googlevideo.com", "ytimg.com", "ggpht.com", "youtube-nocookie.com", "youtubei.googleapis.com", "youtube.googleapis.com", "youtubeembeddedplayer.googleapis.com", "youtubekids.com", "youtu.be", "yt3.googleusercontent.com", "wide-youtube.l.google.com", "youtube-ui.l.google.com", "ytimg.l.google.com", "jnn-pa.googleapis.com"} {
		if DomainMatches(host, suffix) {
			return "youtube-media"
		}
	}
	return ""
}

// DirectTimeoutHost shares failure evidence within an explicit family, while
// ordinary sites remain exact-host scoped. Network/profile/port stay in the key.
func DirectTimeoutHost(host string) string {
	if BrowserRelayGroup(host) != "" {
		return "youtube.com"
	}
	return strings.ToLower(strings.TrimSuffix(host, "."))
}

// SetProbeHost remembers the actual failed origin, so a healthy unrelated member
// cannot accidentally clear the family verdict at TTL. It creates no verdict.
func (c *DirectTimeoutCache) SetProbeHost(k DirectTimeoutKey, host string) {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if len(host) > 253 || BrowserRelayGroup(host) == "" || DirectTimeoutHost(host) != k.Host {
		return
	}
	k.Address = ""
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := c.entries[k]; e != nil {
		e.probeHost = host
	}
}
func (c *DirectTimeoutCache) ProbeHost(k DirectTimeoutKey) string {
	k.Address = ""
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := c.entries[k]; e != nil {
		return e.probeHost
	}
	return ""
}
