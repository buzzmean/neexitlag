package localstrategy

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/netip"
	"sort"
	"strings"
	"sync/atomic"
)

const Version = 1
const MaxProfiles = 32

type Profile struct {
	SplitPosition  uint16   `json:"split_position,omitempty"`
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Version        int      `json:"version"`
	Protocol       string   `json:"protocol"`
	Ports          []uint16 `json:"ports"`
	Application    string   `json:"application"`
	Domains        []string `json:"domains"`
	IPs            []string `json:"ips"`
	ExcludeDomains []string `json:"exclude_domains"`
	ExcludeIPs     []string `json:"exclude_ips"`
	Strategy       string   `json:"strategy"`
	Cutoff         uint32   `json:"cutoff"`
	Priority       int      `json:"priority"`
	Enabled        bool     `json:"enabled"`
}
type Config struct {
	Version int       `json:"version"`
	Enabled bool      `json:"enabled"`
	Builtin []Profile `json:"builtin"`
	User    []Profile `json:"user"`
}
type compiled struct {
	Profile
	domains  DomainSet
	ips      IPSet
	excludes ExcludeSet
}
type snapshot struct {
	config     Config
	profiles   []compiled
	generation uint64
	digest     string
}
type Metadata struct {
	Application, Domain, Protocol string
	StrategyID                    string
	Destination                   netip.AddrPort
}
type State struct {
	selectedIndex int
	splitPosition uint16
	selectedID    string
	generation    uint64
	inspected     uint32
	done, matched bool
	applied       bool
	domain        string
	failed        bool
}
type Result struct {
	Active, First    bool
	Packets          [][]byte
	Strategy, Domain string
	Modified         bool
	Failed           bool
}

// LocalStrategy has no capture handle. One existing NeExitLag capture feeds it.
type LocalStrategy interface {
	Match(Metadata) bool
	Process([]byte, *State) Result
}
type Metrics struct {
	RelaySavingsMeasured bool   `json:"relay_savings_measured"`
	Inspected            uint64 `json:"packets_inspected"`
	Modified             uint64 `json:"packets_modified"`
	Extra                uint64 `json:"extra_packets_generated"`
	Flows                uint64 `json:"flows_direct_plus"`
	Bytes                uint64 `json:"bytes_direct_plus"`
	RelayAvoided         uint64 `json:"bytes_relay_avoided"`
	Errors               uint64 `json:"errors"`
}
type LocalStrategyEngine struct {
	revision                                         atomic.Uint64
	Guard                                            func(Metadata) bool
	suspended                                        atomic.Bool
	current                                          atomic.Pointer[snapshot]
	inspected, modified, extra, flows, bytes, errors atomic.Uint64
}

func New() *LocalStrategyEngine {
	e := &LocalStrategyEngine{}
	initial, _ := build(Config{Version: Version, Builtin: []Profile{}, User: []Profile{}})
	e.current.Store(initial)
	return e
}
func compile(p Profile) (compiled, error) {
	if p.ID == "" || len(p.ID) > 64 || len(p.Name) > 128 || len(p.Application) > 2048 || p.Version != Version || p.Cutoff < 1 || p.Cutoff > 8 || len(p.Ports) == 0 || len(p.Ports) > 32 || !strings.HasSuffix(strings.ToLower(p.Application), ".exe") {
		return compiled{}, errors.New("invalid local strategy profile")
	}
	mechanism, ok := mechanisms[p.Strategy]
	if !ok || !mechanism.Match(Metadata{Protocol: p.Protocol}) {
		return compiled{}, errors.New("unsupported protocol/strategy")
	}
	if p.Protocol == "udp" && (len(p.Domains) > 0 || len(p.ExcludeDomains) > 0) {
		return compiled{}, errors.New("UDP domain metadata unsupported; use explicit IP/exclude IP")
	}

	if p.Strategy == "tcp-split" {
		if p.SplitPosition < 1 || p.SplitPosition > 512 {
			return compiled{}, errors.New("TCP split position must be 1..512")
		}
	} else if p.SplitPosition != 0 {
		return compiled{}, errors.New("split position requires tcp-split mechanism")
	}
	for _, port := range p.Ports {
		if port == 0 {
			return compiled{}, errors.New("invalid port")
		}
	}
	if len(p.Domains)+len(p.IPs) == 0 {
		return compiled{}, errors.New("explicit domain/IP target required")
	}
	if p.Protocol == "udp" && len(p.IPs) == 0 {
		return compiled{}, errors.New("UDP requires explicit IP set; QUIC SNI is not inferred")
	}
	c := compiled{Profile: p}
	var err error
	if c.domains, err = NewDomainSet(p.Domains); err != nil {
		return c, err
	}
	if c.ips, err = NewIPSet(p.IPs); err != nil {
		return c, err
	}
	if c.excludes.Domains, err = NewDomainSet(p.ExcludeDomains); err != nil {
		return c, err
	}
	c.excludes.IPs, err = NewIPSet(p.ExcludeIPs)
	return c, err
}
func (e *LocalStrategyEngine) Validate(c Config) error { _, err := build(c); return err }
func build(c Config) (*snapshot, error) {
	raw, err := json.Marshal(c)
	if err != nil || len(raw) > 24000 || c.Version != Version || len(c.Builtin)+len(c.User) > MaxProfiles {
		return nil, errors.New("invalid/bounded local strategy config")
	}
	var copy Config
	json.Unmarshal(raw, &copy)
	if copy.Builtin == nil {
		copy.Builtin = []Profile{}
	}
	if copy.User == nil {
		copy.User = []Profile{}
	}
	raw, _ = json.Marshal(copy)
	hash := sha256.Sum256(raw)
	s := &snapshot{config: copy, digest: hex.EncodeToString(hash[:])}
	ids := map[string]bool{}
	for _, p := range append(append([]Profile{}, copy.Builtin...), copy.User...) {
		if ids[p.ID] {
			return nil, errors.New("duplicate strategy id")
		}
		ids[p.ID] = true
		v, err := compile(p)
		if err != nil {
			return nil, err
		}
		s.profiles = append(s.profiles, v)
	}
	sort.SliceStable(s.profiles, func(i, j int) bool { return s.profiles[i].Priority > s.profiles[j].Priority })
	return s, nil
}

// Reload publishes an immutable complete generation. Existing state stops further
// transforms on reload; it never reclassifies a live connection to another route.
func (e *LocalStrategyEngine) Reload(c Config) error {
	s, err := build(c)
	if err != nil {
		return err
	}
	s.generation = e.revision.Add(1)
	e.current.Store(s)
	return nil
}
func (e *LocalStrategyEngine) Config() Config {
	b, _ := json.Marshal(e.current.Load().config)
	var c Config
	json.Unmarshal(b, &c)
	return c
}
func appMatch(pattern, path string) bool {
	p := strings.ToLower(strings.ReplaceAll(pattern, "\\", "/"))
	v := strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
	if !strings.Contains(p, "/") {
		if n := strings.LastIndex(v, "/"); n >= 0 {
			v = v[n+1:]
		}
	}
	return p == v
}
func (c *compiled) Match(m Metadata) bool {
	if !c.Enabled || c.Protocol != m.Protocol || !appMatch(c.Application, m.Application) || c.excludes.Contains(m.Domain, m.Destination.Addr()) {
		return false
	}
	port := false
	for _, p := range c.Ports {
		port = port || p == m.Destination.Port()
	}
	if !port {
		return false
	}
	if len(c.ExcludeDomains) > 0 && m.Domain == "" {
		return false
	}
	return c.ips.Contains(m.Destination.Addr()) || c.domains.Contains(m.Domain)
}
func excluded(s *snapshot, m Metadata) bool {
	for i := range s.profiles {
		c := &s.profiles[i]
		if !c.Enabled || c.Protocol != m.Protocol || !appMatch(c.Application, m.Application) {
			continue
		}
		port := false
		for _, p := range c.Ports {
			port = port || p == m.Destination.Port()
		}
		if port && (c.excludes.Contains(m.Domain, m.Destination.Addr()) || len(c.ExcludeDomains) > 0 && m.Domain == "") {
			return true
		}
	}
	return false
}
func (e *LocalStrategyEngine) Match(m Metadata) bool {
	s := e.current.Load()
	if !s.config.Enabled || excluded(s, m) {
		return false
	}
	for i := range s.profiles {
		if (m.StrategyID == "" || s.profiles[i].ID == m.StrategyID) && s.profiles[i].Match(m) {
			return true
		}
	}
	return false
}

// Process is bounded to the first Cutoff payload packets. No payload is retained.
// State belongs to an existing bounded capture flow, so no extra connection table.
func (e *LocalStrategyEngine) Process(p []byte, st *State, m Metadata) (r Result) {
	defer func() {
		if recover() != nil {
			st.Disable()
			e.errors.Add(1)
			r = Result{Failed: true}
		}
	}()
	if st == nil {
		e.errors.Add(1)
		return Result{Failed: true}
	}
	s := e.current.Load()
	if e.suspended.Load() || !s.config.Enabled || st.generation != 0 && st.generation != s.generation {
		st.Disable()
		return Result{}
	}
	if st.done {
		if st.applied {
			return Result{Strategy: strategyID(st), Domain: st.domain, Active: true}
		}
		return Result{}
	}
	payload, ok := packetPayload(p, m.Protocol)
	if !ok {
		return Result{}
	}
	if len(payload) == 0 {
		return Result{}
	}
	var destination [4]byte
	copy(destination[:], p[16:20])
	if netip.AddrPortFrom(netip.AddrFrom4(destination), binary.BigEndian.Uint16(p[22:24])) != m.Destination {
		st.done = true
		return Result{}
	}
	st.generation = s.generation
	st.inspected++
	e.inspected.Add(1)
	if !st.matched {
		if m.Protocol == "tcp" {
			m.Domain = PacketDomain(payload)
		}
		if excluded(s, m) {
			st.done = true
			return Result{}
		}
		for i := range s.profiles {
			if (m.StrategyID == "" || s.profiles[i].ID == m.StrategyID) && s.profiles[i].Match(m) {
				st.selectedIndex = i
				st.selectedID = s.profiles[i].ID
				if e.Guard != nil {
					m.StrategyID = st.selectedID
					if !e.Guard(m) {
						st.done = true
						return Result{}
					}
				}
				st.matched = true
				break
			}
		}
		// No reassembly or guessing when metadata is absent; never modify midstream.
		if !st.matched {
			st.done = true
			return Result{}
		}
	}
	c := &s.profiles[st.selectedIndex]
	r.Strategy = c.ID
	if st.domain == "" {
		st.domain = m.Domain
	}
	r.Domain = st.domain
	if st.inspected >= c.Cutoff {
		st.done = true
	}
	st.splitPosition = c.SplitPosition
	processed := mechanisms[c.Strategy].Process(p, st)
	r.Packets = processed.Packets
	r.Modified = processed.Modified
	if r.Modified {
		r.First = !st.applied
		st.applied = true
	}
	r.Active = st.applied
	return r
}
func strategyID(s *State) string { return s.selectedID }

// Commit is called only after successful reinjection. Wire counters do not imply
// application success. Avoided relay bytes are deliberately not estimated.
func (e *LocalStrategyEngine) Commit(r Result, n int) {
	if r.First && r.Modified {
		e.flows.Add(1)
	}
	if r.Modified {
		e.modified.Add(1)
		e.extra.Add(uint64(len(r.Packets) - 1))
	}
	if r.Active {
		e.bytes.Add(uint64(n))
	}
}
func (e *LocalStrategyEngine) Metrics() Metrics {
	return Metrics{Inspected: e.inspected.Load(), Modified: e.modified.Load(), Extra: e.extra.Load(), Flows: e.flows.Load(), Bytes: e.bytes.Load(), Errors: e.errors.Load()}
}

func (s *State) Disable() {
	s.done = true
	s.failed = true
	s.matched = false
	s.selectedID = ""
	s.applied = false
	s.domain = ""
}
func (s *State) Applied() bool { return s.applied && !s.failed }

func (e *LocalStrategyEngine) Suspend(v bool) { e.suspended.Store(v) }

func (e *LocalStrategyEngine) Revision() string { return e.current.Load().digest }

func (e *LocalStrategyEngine) Fail(s *State) { s.Disable(); e.errors.Add(1) }
