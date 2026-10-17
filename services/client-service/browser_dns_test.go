package clientservice

import (
	"context"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"testing"
)

func TestOriginComparisonStrictlyRelayOnlyAndReadOnly(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.Profiles.Upsert(profiles.Profile{ID: "keep", Name: "existing", Executables: []string{`C:\Apps\chrome.exe`}, Mode: profiles.Auto})
	before, _ := json.Marshal(c.Profiles.List())
	for _, raw := range []string{`{"host":"a.local","candidates":["8.8.8.8"]}`, `{"host":"example.com","candidates":["127.0.0.1"]}`, `{"host":"example.com","candidates":["169.254.169.254"]}`, `{"host":"example.com","candidates":["8.8.8.8","8.8.8.8"]}`, `{"host":"example.com","candidates":[]}`, `{"host":"example.com","candidates":["8.8.8.8"],"insecure":true}`} {
		if _, err := c.Handle(context.Background(), "browser.origin.check", []byte(raw)); err == nil {
			t.Fatal("invalid diagnostic accepted", raw)
		}
	}
	// No relay configured: a public destination must fail BEFORE any DIRECT dial.
	response, err := c.Handle(context.Background(), "browser.origin.check", []byte(`{"host":"example.com","candidates":["8.8.8.8"]}`))
	if err != nil {
		t.Fatal(err)
	}
	wire, _ := json.Marshal(response)
	var report struct {
		Candidates []struct {
			Route    string                          `json:"route"`
			Verified bool                            `json:"origin_verified"`
			Origin   struct{ Failure, Phase string } `json:"origin"`
		} `json:"candidates"`
	}
	if json.Unmarshal(wire, &report) != nil || len(report.Candidates) != 1 {
		t.Fatal(string(wire))
	}
	result := report.Candidates[0]
	if result.Route != string(flowpolicy.RelayRequired) || result.Verified || result.Origin.Phase != "RELAY_OR_TCP_OPEN" || result.Origin.Failure == "" {
		t.Fatal(string(wire))
	}
	after, _ := json.Marshal(c.Profiles.List())
	if string(before) != string(after) || c.Snapshot().Optimizing {
		t.Fatal("read-only check changed profiles/optimization")
	}
}
