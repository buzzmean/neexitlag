package windowsplatform

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func validEdgeProxyValue(raw string) bool {
	if len(raw) > 256*1024 {
		return false
	}
	var p struct {
		Mode      string `json:"ProxyMode"`
		URL       string `json:"ProxyPacUrl"`
		Mandatory bool   `json:"ProxyPacMandatory"`
	}
	if json.Unmarshal([]byte(raw), &p) != nil || p.Mode != "pac_script" || !p.Mandatory {
		return false
	}
	const prefix = "data:application/x-ns-proxy-autoconfig;base64,"
	if !strings.HasPrefix(p.URL, prefix) || len(p.URL) > 128*1024 {
		return false
	}
	b, e := base64.StdEncoding.DecodeString(strings.TrimPrefix(p.URL, prefix))
	return e == nil && len(b) > 0 && strings.Contains(string(b), "FindProxyForURL")
}
