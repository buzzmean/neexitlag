package clientservice

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/security"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
)

// Optional v1 query limits desktop results to configured applications instead of
// overflowing the IPC frame with unrelated system processes. Old nil queries work.
func listProcesses(raw json.RawMessage) ([]windowsplatform.ProcessIdentity, error) {
	var query struct {
		Executables []string `json:"executables"`
	}
	if len(raw) > 0 && string(raw) != "null" {
		if err := security.Parse(raw, &query); err != nil {
			return nil, err
		}
	}
	if len(query.Executables) > 2048 {
		return nil, errors.New("process query limit exceeded")
	}
	for _, exe := range query.Executables {
		if len(exe) > 2048 || !strings.HasSuffix(strings.ToLower(exe), ".exe") {
			return nil, errors.New("invalid executable query")
		}
	}
	all, err := windowsplatform.NewDiscovery().List()
	if err != nil {
		return nil, err
	}
	return filterProcesses(all, query.Executables), nil
}
func filterProcesses(all []windowsplatform.ProcessIdentity, executables []string) []windowsplatform.ProcessIdentity {
	if len(executables) == 0 {
		return all
	}
	match := profiles.Profile{Executables: executables}
	result := []windowsplatform.ProcessIdentity{}
	for _, id := range all {
		if match.Matches(id) {
			result = append(result, id)
		}
	}
	return result
}
