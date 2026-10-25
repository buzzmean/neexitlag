package clientservice

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/buzzmean/neexitlag/core/release"
	"github.com/buzzmean/neexitlag/core/security"
)

// BuildSequence is the git commit count of this build (set with -ldflags -X).
// "0" (developer build) never auto-updates.
var BuildSequence = "0"

const updateRepository = "buzzmean/veliq"
const updateInterval = 6 * time.Hour

type UpdateStatus struct {
	State     string    `json:"state"`
	Current   string    `json:"current_build"`
	Available uint64    `json:"available_build,omitempty"`
	Version   string    `json:"available_version,omitempty"`
	Error     string    `json:"error,omitempty"`
	Checked   time.Time `json:"checked_at,omitempty"`
}

type updater struct {
	mu      sync.Mutex
	status  UpdateStatus
	http    *http.Client
	keys    map[string]ed25519.PublicKey
	dir     string // service install directory, e.g. %ProgramData%\VELIQ
	api     string
	launch  func(path string) error
	current uint64
}

func newUpdater(keys map[string]ed25519.PublicKey, dir string, launch func(string) error) *updater {
	current, _ := strconv.ParseUint(BuildSequence, 10, 64)
	return &updater{status: UpdateStatus{State: "IDLE", Current: BuildSequence}, http: &http.Client{Timeout: 30 * time.Minute}, keys: keys, dir: dir, api: "https://api.github.com/repos/" + updateRepository + "/releases/latest", launch: launch, current: current}
}

func (u *updater) snapshot() UpdateStatus {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.status
}

func (u *updater) set(state, failure string, m *release.Manifest) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.status.State, u.status.Error, u.status.Checked = state, failure, time.Now()
	if m != nil {
		u.status.Available, u.status.Version = m.Build, m.Version
	}
}

// Club images are frozen and diskless: an update would be downloaded again on
// every boot and lost. The owner updates the image instead.
func (u *updater) disabled() string {
	if u.current == 0 {
		return "DEVELOPER_BUILD"
	}
	if _, e := os.Stat(filepath.Join(u.dir, "club-mode.json")); e == nil {
		return "CLUB_IMAGE_MANAGED"
	}
	return ""
}

func (u *updater) run(ctx context.Context) {
	timer := time.NewTimer(10 * time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		u.check(ctx)
		timer.Reset(updateInterval)
	}
}

type githubRelease struct {
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

func (u *updater) check(ctx context.Context) {
	if reason := u.disabled(); reason != "" {
		u.set("DISABLED", reason, nil)
		return
	}
	m, setupURL, e := u.latest(ctx)
	if e != nil {
		u.set("CHECK_FAILED", e.Error(), nil)
		return
	}
	if m.Build <= u.current {
		u.set("UP_TO_DATE", "", &m)
		return
	}
	if u.recentlyAttempted(m.Build) {
		u.set("WAITING_RETRY", "previous install of this build did not complete", &m)
		return
	}
	u.set("DOWNLOADING", "", &m)
	path, e := u.download(ctx, setupURL, m)
	if e != nil {
		u.set("DOWNLOAD_FAILED", e.Error(), &m)
		return
	}
	u.markAttempt(m.Build)
	u.set("INSTALLING", "", &m)
	// Setup replaces and restarts this service; it is not waited for.
	if e = u.launch(path); e != nil {
		u.set("INSTALL_FAILED", e.Error(), &m)
	}
}

func (u *updater) get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "VELIQ-updater")
	resp, e := u.http.Do(req)
	if e != nil {
		return nil, errors.New("UPDATE_SOURCE_UNREACHABLE")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("UPDATE_SOURCE_HTTP_%d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

func (u *updater) latest(ctx context.Context) (release.Manifest, string, error) {
	var m release.Manifest
	body, e := u.get(ctx, u.api, 1<<20)
	if e != nil {
		return m, "", e
	}
	var r githubRelease
	if json.Unmarshal(body, &r) != nil {
		return m, "", errors.New("UPDATE_RELEASE_INVALID")
	}
	manifestURL, setupURL := "", ""
	for _, a := range r.Assets {
		switch a.Name {
		case release.ManifestAsset:
			manifestURL = a.URL
		case release.SetupAsset:
			setupURL = a.URL
		}
	}
	if manifestURL == "" || setupURL == "" {
		return m, "", errors.New("UPDATE_ASSETS_MISSING")
	}
	signed, e := u.get(ctx, manifestURL, security.MaxDocument)
	if e != nil {
		return m, "", e
	}
	if security.Verify(release.Domain, u.keys, signed, &m) != nil || m.Validate() != nil {
		return m, "", errors.New("UPDATE_SIGNATURE_INVALID")
	}
	return m, setupURL, nil
}

func (u *updater) download(ctx context.Context, url string, m release.Manifest) (string, error) {
	dir := filepath.Join(u.dir, "updates")
	if e := os.MkdirAll(dir, 0o700); e != nil {
		return "", e
	}
	path := filepath.Join(dir, fmt.Sprintf("VELIQ-Setup-%d.exe", m.Build))
	if ok, _ := fileMatches(path, m); ok {
		return path, nil
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return "", e
	}
	req.Header.Set("User-Agent", "VELIQ-updater")
	resp, e := u.http.Do(req)
	if e != nil {
		return "", errors.New("UPDATE_SOURCE_UNREACHABLE")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("UPDATE_SOURCE_HTTP_%d", resp.StatusCode)
	}
	tmp := path + ".part"
	f, e := os.Create(tmp)
	if e != nil {
		return "", e
	}
	h := sha256.New()
	n, e := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, m.SetupSize+1))
	if ce := f.Close(); e == nil {
		e = ce
	}
	if e != nil || n != m.SetupSize || hex.EncodeToString(h.Sum(nil)) != m.SetupSHA256 {
		os.Remove(tmp)
		return "", errors.New("UPDATE_DOWNLOAD_MISMATCH")
	}
	if e = os.Rename(tmp, path); e != nil {
		return "", e
	}
	return path, nil
}

func fileMatches(path string, m release.Manifest) (bool, error) {
	f, e := os.Open(path)
	if e != nil {
		return false, e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, f)
	return e == nil && n == m.SetupSize && hex.EncodeToString(h.Sum(nil)) == m.SetupSHA256, e
}

// One automatic attempt per build per day: a failing Setup must not loop.
func (u *updater) attemptFile() string { return filepath.Join(u.dir, "updates", "attempt.json") }
func (u *updater) recentlyAttempted(build uint64) bool {
	raw, e := os.ReadFile(u.attemptFile())
	if e != nil {
		return false
	}
	var a struct {
		Build uint64    `json:"build"`
		At    time.Time `json:"at"`
	}
	return json.Unmarshal(raw, &a) == nil && a.Build == build && time.Since(a.At) < 24*time.Hour
}
func (u *updater) markAttempt(build uint64) {
	raw, _ := json.Marshal(map[string]any{"build": build, "at": time.Now().UTC()})
	_ = os.WriteFile(u.attemptFile(), raw, 0o600)
}
