package browserfixture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (f *Fixture) RunChrome(ctx context.Context, chrome string, args []string, transitions ...func() error) (string, error) {
	directory := ""
	for _, a := range args {
		if strings.HasPrefix(a, "--user-data-dir=") {
			directory = strings.TrimPrefix(a, "--user-data-dir=")
		}
	}
	if directory == "" {
		return "", errors.New("isolated browser data directory required")
	}
	args = append(args, "--headless=new", "--disable-background-networking", "--disable-component-update", "--disable-sync", "--disable-domain-reliability", "--disable-default-apps", "--dns-prefetch-disable", "--disable-features=MediaRouter,OptimizationHints", "--no-default-browser-check", "--password-store=basic", "--use-mock-keychain", "--disable-dev-shm-usage", "--ignore-certificate-errors-spki-list="+f.SPKI, "--remote-debugging-port=0", "--remote-allow-origins=http://localhost", "about:blank")
	// Linux containers run as root. Native Windows must exercise the actual
	// Chrome sandbox/token owner path, not a weakened --no-sandbox fixture.
	if runtime.GOOS != "windows" {
		args = append(args, "--no-sandbox")
	}
	cmd := exec.CommandContext(ctx, chrome, args...)
	cmd.Stdout = io.Discard
	stderr := &chromeLog{}
	cmd.Stderr = stderr
	if e := cmd.Start(); e != nil {
		return "", e
	}
	ended := make(chan error, 1)
	go func() { ended <- cmd.Wait() }()
	defer func() {
		cmd.Process.Kill()
		select {
		case <-ended:
		case <-time.After(3 * time.Second):
		}
	}()
	var port int
	ready, stopReady := context.WithTimeout(ctx, 25*time.Second)
	defer stopReady()
	for {
		raw, e := os.ReadFile(filepath.Join(directory, "DevToolsActivePort"))
		if e == nil {
			lines := strings.Split(string(raw), "\n")
			port, _ = strconv.Atoi(lines[0])
			if port > 0 {
				break
			}
		}
		select {
		case exited := <-ended:
			ended <- exited
			return "", fmt.Errorf("CDP/INSTRUMENTATION_FAILURE: Chrome exited: %v; %s", exited, stderr.String())
		case <-ready.Done():
			return "", fmt.Errorf("CDP/INSTRUMENTATION_FAILURE: port readiness: %w; %s", ready.Err(), stderr.String())
		case <-time.After(100 * time.Millisecond):
		}
	}
	if port == 0 {
		return "", fmt.Errorf("CDP/INSTRUMENTATION_FAILURE: missing port; %s", stderr.String())
	}
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	pageURL, e := waitCDPPage(ctx, client, fmt.Sprintf("http://127.0.0.1:%d/json/list", port))
	if e != nil {
		return "", fmt.Errorf("%w; Chrome stderr: %s", e, stderr.String())
	}
	conn, e := websocket.Dial(pageURL, "", "http://localhost")
	if e != nil {
		return "", e
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}
	sequence := 0
	call := func(method string, params any) (json.RawMessage, error) {
		sequence++
		if e := websocket.JSON.Send(conn, map[string]any{"id": sequence, "method": method, "params": params}); e != nil {
			return nil, e
		}
		for {
			var message struct {
				ID     int             `json:"id"`
				Result json.RawMessage `json:"result"`
				Error  json.RawMessage `json:"error"`
			}
			if e := websocket.JSON.Receive(conn, &message); e != nil {
				return nil, e
			}
			if message.ID == sequence {
				if len(message.Error) > 0 {
					return nil, errors.New("CDP instrumentation rejected command")
				}
				return message.Result, nil
			}
		}
	}
	if f.HealthURL != "" {
		if _, err := call("Page.navigate", map[string]any{"url": f.HealthURL}); err != nil {
			return "", err
		}
		observed := false
		for i := 0; i < 100; i++ {
			raw, err := call("Runtime.evaluate", map[string]any{"expression": "document.body&&document.body.innerText", "returnByValue": true})
			if err != nil {
				return "", err
			}
			var r struct {
				Result struct {
					Value string `json:"value"`
				} `json:"result"`
			}
			json.Unmarshal(raw, &r)
			if strings.Contains(r.Result.Value, "NeExitLag Chrome transport OK") {
				observed = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if !observed {
			return "", errors.New("actual Chrome PAC/owner readiness failed")
		}
	}
	if f.BeforeRequests != nil {
		if err := f.BeforeRequests(); err != nil {
			return "", err
		}
	}
	title := ""
	for stage := 1; stage <= len(transitions)+1; stage++ {
		if stage > 1 {
			if e := transitions[stage-2](); e != nil {
				return title, e
			}
		}
		expected := "NEEXITLAG_PASS_" + f.Nonce + "_" + strconv.Itoa(stage)
		_, e = call("Page.navigate", map[string]any{"url": fmt.Sprintf("https://ordinary.example:%d/bootstrap?stage=%d", f.Port(), stage)})
		if e != nil {
			return "", e
		}
		title = ""
		for tries := 0; tries < 200; tries++ {
			raw, e := call("Runtime.evaluate", map[string]any{"expression": "document.title", "returnByValue": true})
			if e != nil {
				return "", e
			}
			var result struct {
				Result struct {
					Value string `json:"value"`
				} `json:"result"`
			}
			if e = json.Unmarshal(raw, &result); e != nil {
				return "", e
			}
			title = result.Result.Value
			if title == expected {
				break
			}
			if title == "NEEXITLAG_FAIL" {
				raw, _ := call("Runtime.evaluate", map[string]any{"expression": "document.body.innerText", "returnByValue": true})
				return title, fmt.Errorf("controlled Chrome stage %d fetch failed: %s", stage, raw)
			}
			select {
			case <-ctx.Done():
				return title, ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
		if title != expected {
			return title, errors.New("Chrome controlled HTTPS result missing")
		}

	}
	// This is an isolated test browser only. Production never opens CDP.
	websocket.JSON.Send(conn, map[string]any{"id": sequence + 1, "method": "Browser.close"})
	select {
	case e := <-ended:
		ended <- e
	case <-time.After(3 * time.Second):
		return title, errors.New("isolated Chrome did not exit")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, host := range Domains {
		if f.Seen[host] == 0 {
			return title, errors.New("controlled target saw no request for an origin")
		}
		if f.Protocols[host] != 2 {
			return title, fmt.Errorf("controlled Chrome origin %s did not negotiate HTTP/2", host)
		}
	}
	return title, nil
}

// DevToolsActivePort is written before the HTTP endpoint/page target is ready.
// Retry instrumentation startup only; HTTP fixture failures are never retried or
// converted to PASS, and the original browser process is never replaced.
func waitCDPPage(parent context.Context, client *http.Client, url string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	for {
		request, e := http.NewRequestWithContext(ctx, "GET", url, nil)
		if e != nil {
			return "", e
		}
		response, e := client.Do(request)
		if e == nil {
			var pages []struct {
				Type string `json:"type"`
				WS   string `json:"webSocketDebuggerUrl"`
			}
			if response.StatusCode == 200 {
				e = json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(&pages)
			}
			response.Body.Close()
			if e == nil {
				for _, p := range pages {
					if p.Type == "page" && p.WS != "" {
						return p.WS, nil
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return "", errors.New("CDP/INSTRUMENTATION_FAILURE: page target readiness deadline")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// Bounded diagnostics for this ephemeral acceptance browser only. Production
// never opens CDP or captures browser stderr/page data.
type chromeLog struct {
	mu   sync.Mutex
	tail []byte
}

func (l *chromeLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(p)
	l.tail = append(l.tail, p...)
	if len(l.tail) > 16384 {
		l.tail = append([]byte(nil), l.tail[len(l.tail)-16384:]...)
	}
	return n, nil
}
func (l *chromeLog) String() string { l.mu.Lock(); defer l.mu.Unlock(); return string(l.tail) }

// RunPolicyHealthProbe mirrors the GUI's empty readiness profile. It supplies
// no PAC/proxy flags: native Windows must use the actual current-user policy.
// Real user tabs are launched only after this process exits and AUTO commits.
func RunPolicyHealthProbe(parent context.Context, chrome, healthURL, directory string) error {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	args := []string{"--headless=new", "--dump-dom", "--timeout=15000", "--no-first-run", "--user-data-dir=" + directory, healthURL}
	if runtime.GOOS != "windows" {
		args = append(args, "--no-sandbox")
	}
	cmd := exec.CommandContext(ctx, chrome, args...)
	stdout, stderr := &chromeLog{}, &chromeLog{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Chrome policy health process: %w; %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "NeExitLag Chrome transport OK") {
		return errors.New("Chrome policy did not reach authenticated readiness endpoint")
	}
	return nil
}
