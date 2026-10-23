package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingIdentityIsNeverCreatedOrEnrolled(t *testing.T) {
	root := t.TempDir()
	identity := filepath.Join(root, "missing.json")
	manifestPath := filepath.Join(root, "service.json")
	os.WriteFile(manifestPath, []byte(`{"device_identity_file":"`+strings.ReplaceAll(identity, `\`, `\\`)+`"}`), 0600)
	if _, _, e := connect(context.Background(), manifestPath, "VPS-01", "146.0.77.98:4433"); e == nil || e.Error() != "EXISTING_IDENTITY_REQUIRED" {
		t.Fatal(e)
	}
	if _, e := os.Stat(identity); !os.IsNotExist(e) {
		t.Fatal("identity created")
	}
}
