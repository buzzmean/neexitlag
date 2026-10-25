package browserproxy

import (
	"encoding/base64"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestAvailabilityPACExplicitDIRECTForEveryHost(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node required")
	}
	args, err := ChromeArgumentsWithRecovery("127.0.0.1:1234", "fixture", []string{"private.example"}, false, "127.0.0.1:1235")
	if err != nil {
		t.Fatal(err)
	}
	pac, err := AvailabilityPAC(strings.TrimPrefix(args[1], "--proxy-pac-url="))
	if err != nil {
		t.Fatal(err)
	}
	again, err := AvailabilityPAC(pac)
	if err != nil || again != pac {
		t.Fatal("not idempotent")
	}
	raw, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(pac, "data:application/x-ns-proxy-autoconfig;base64,"))
	input, _ := json.Marshal(map[string]any{"pac": string(raw), "names": []string{"chatgpt.com", "www.youtube.com", "a.googlevideo.com", "private.example", "ordinary.example"}})
	cmd := exec.Command(node, "-e", `let d='';process.stdin.on('data',v=>d+=v);process.stdin.on('end',()=>{let x=JSON.parse(d),c={};require('vm').runInNewContext(x.pac,c);process.stdout.write(JSON.stringify(x.names.map(n=>c.FindProxyForURL('https://'+n,n))));});`)
	cmd.Stdin = strings.NewReader(string(input))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	var routes []string
	if json.Unmarshal(out, &routes) != nil || len(routes) != 5 {
		t.Fatal("PAC execution failed")
	}
	for i, v := range routes {
		if !strings.HasSuffix(v, "; DIRECT") {
			t.Fatal("DIRECT absent", i)
		}
	}
	if strings.Contains(routes[0], ":1235") {
		t.Fatal("protected origin acquired secondary helper")
	}
	if !strings.Contains(routes[4], ":1235") {
		t.Fatal("ordinary secondary missing")
	}
}
