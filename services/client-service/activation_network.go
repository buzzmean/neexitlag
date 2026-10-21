package clientservice

import (
	"context"
	"errors"
	"github.com/buzzmean/neexitlag/core/enrollment"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

// New/unconfigured controllers and valid legacy installations retain their
// established behavior. Code-managed boot must validate binding/authentication
// before any WinDivert handle or browser policy is activated.
func (c *Controller) activationNetworkAllowed() bool { return !c.activation.networkBlocked.Load() }
func (c *Controller) activationHardware(ctx context.Context, salt []byte) (string, error) {
	c.activation.mu.Lock()
	hardware := c.activation.hardware
	c.activation.mu.Unlock()
	if hardware == nil {
		hardware = wp.EnrollmentHardware
	}
	h, err := hardware(ctx)
	hash, report, hashErr := enrollment.HardwareHashDiagnostics(h, salt)
	if err != nil {
		hash = ""
		report = enrollment.HardwareDiagnostics{UUID: "NOT_QUERIED", BoardSerial: "NOT_QUERIED", MAC: "NOT_QUERIED", Reason: "HARDWARE_METADATA_QUERY_FAILED"}
		hashErr = errors.New("HARDWARE_ID_UNAVAILABLE")
	}
	c.activation.mu.Lock()
	c.activation.diagnostics = report
	c.activation.mu.Unlock()
	return hash, hashErr
}
func activationNetworkMutation(method string) bool {
	switch method {
	case "browser.native.enable", "browser.prepare", "browser.start", "browser.confirm", "browser.silent.apply", "browser.silent.guard", "edge.prepare", "edge.confirm", "edge.policy", "edge.silent.apply":
		return true
	}
	return false
}
