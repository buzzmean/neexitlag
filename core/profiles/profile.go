package profiles

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Mode string

const (
	Auto   Mode = "AUTO"
	Direct Mode = "DIRECT"
	Manual Mode = "MANUAL"
	Block  Mode = "BLOCK"
)

type Profile struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Executables     []string `json:"executables"`
	Publisher       string   `json:"publisher,omitempty"`
	SignatureSHA256 string   `json:"signature_sha256,omitempty"`
	ProtocolHints   []string `json:"protocol_hints,omitempty"`
	PortHints       []uint16 `json:"port_hints,omitempty"`
	Mode            Mode     `json:"mode"`
	RelayID         string   `json:"relay_id,omitempty"`
	Custom          bool     `json:"custom"`
}

func (p Profile) Validate() error {
	if len(p.ID) < 1 || len(p.ID) > 128 || p.Name == "" || len(p.Name) > 128 || len(p.Executables) < 1 || len(p.Executables) > 16 {
		return errors.New("invalid application profile")
	}
	for _, exe := range p.Executables {
		if len(exe) > 2048 || !strings.HasSuffix(strings.ToLower(exe), ".exe") {
			return errors.New("profile requires .exe identity")
		}
	}
	switch p.Mode {
	case Auto, Direct, Block:
	case Manual:
		if p.RelayID == "" {
			return errors.New("manual mode requires relay")
		}
	default:
		return errors.New("invalid routing mode")
	}
	return nil
}
func Defaults() []Profile {
	return []Profile{{ID: "cs2", Name: "Counter-Strike 2", Executables: []string{"cs2.exe"}, Mode: Auto, ProtocolHints: []string{"udp", "tcp"}}, {ID: "discord", Name: "Discord", Executables: []string{"Discord.exe"}, Mode: Auto}, {ID: "steam", Name: "Steam", Executables: []string{"steam.exe"}, Mode: Direct}}
}

type Identity struct {
	PID                                    uint32
	Path, Name, Publisher, SignatureSHA256 string
	CreationTime                           uint64
}

func normalize(s string) string { return strings.ToLower(strings.ReplaceAll(s, "\\", "/")) }
func (p Profile) Matches(i Identity) bool {
	if i.PID == 0 || i.Path == "" || i.CreationTime == 0 {
		return false
	}
	if p.Publisher != "" && p.Publisher != i.Publisher {
		return false
	}
	if p.SignatureSHA256 != "" && !strings.EqualFold(p.SignatureSHA256, i.SignatureSHA256) {
		return false
	}
	path := normalize(i.Path)
	for _, exe := range p.Executables {
		v := normalize(exe)
		if strings.Contains(v, "/") {
			if v == path {
				return true
			}
		} else if v == filepath.Base(path) {
			return true
		}
	}
	return false
}

type Registry struct {
	mu    sync.RWMutex
	items map[string]Profile
}

func New(all []Profile) *Registry {
	r := &Registry{items: make(map[string]Profile)}
	for _, p := range all {
		r.Upsert(p)
	}
	return r
}
func (r *Registry) Upsert(p Profile) error {
	if e := p.Validate(); e != nil {
		return e
	}
	p.Executables = append([]string(nil), p.Executables...)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[p.ID]; !ok && len(r.items) >= 128 {
		return errors.New("profile capacity reached")
	}
	total := 0
	for id, old := range r.items {
		if id != p.ID {
			b, _ := json.Marshal(old)
			total += len(b)
		}
	}
	b, _ := json.Marshal(p)
	if total+len(b) > 24000 {
		return errors.New("profile storage limit reached")
	}
	r.items[p.ID] = p
	return nil
}
func (r *Registry) List() []Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Profile, 0, len(r.items))
	for _, p := range r.items {
		p.Executables = append([]string(nil), p.Executables...)
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Registry) Match(i Identity) (Profile, bool) {
	all := r.List()
	sort.SliceStable(all, func(a, b int) bool { return all[a].Custom && !all[b].Custom })
	for _, p := range all {
		if p.Matches(i) {
			return p, true
		}
	}
	return Profile{}, false
}

func (r *Registry) Delete(id string) { r.mu.Lock(); defer r.mu.Unlock(); delete(r.items, id) }
