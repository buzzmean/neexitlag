package browserproxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

// PAC is inline: it has no fetch failure that could silently revert to DIRECT.
// Every origin has a cryptographically distinct ProxyServer hostname, so Chrome
// cannot coalesce a protected origin into another origin's DIRECT HTTP/2 tunnel.
// Chrome resolves these synthetic *.localhost proxy names internally to
// loopback. Windows DNS/proxy settings are never modified.
func PAC(port uint16) string { return guardedPAC(port, nil, false) }

// Explicit policy domains and YouTube never get a DIRECT proxy failover.
// Only a dead local proxy permits ordinary origins to use their normal path;
// a 403/502 from a running proxy is not a reason to bypass route policy.
func guardedPAC(port uint16, extra []string, strict bool) string {
	guard := []string{"chatgpt.com", "openai.com", "oaistatic.com", "oaiusercontent.com", "chat.com", "youtube.com", "youtu.be", "googlevideo.com", "ytimg.com", "youtube-nocookie.com", "ggpht.com"}
	guard = append(guard, extra...)
	raw, _ := json.Marshal(guard)
	prefix := fmt.Sprintf("var protectedSuffixes=%s; var strictProxy=%t;\n", raw, strict)
	return prefix + fmt.Sprintf(`function sha256(s) {
 var k=[1116352408,1899447441,3049323471,3921009573,961987163,1508970993,2453635748,2870763221,3624381080,310598401,607225278,1426881987,1925078388,2162078206,2614888103,3248222580,3835390401,4022224774,264347078,604807628,770255983,1249150122,1555081692,1996064986,2554220882,2821834349,2952996808,3210313671,3336571891,3584528711,113926993,338241895,666307205,773529912,1294757372,1396182291,1695183700,1986661051,2177026350,2456956037,2730485921,2820302411,3259730800,3345764771,3516065817,3600352804,4094571909,275423344,430227734,506948616,659060556,883997877,958139571,1322822218,1537002063,1747873779,1955562222,2024104815,2227730452,2361852424,2428436474,2756734187,3204031479,3329325298];
 var h=[1779033703,3144134277,1013904242,2773480762,1359893119,2600822924,528734635,1541459225], a=[],i,j,n=s.length*8;
 for(i=0;i<s.length;i++)a.push(s.charCodeAt(i));a.push(128);while(a.length%%64!==56)a.push(0);for(i=7;i>=0;i--)a.push(i<4?(n>>>(i*8))&255:0);
 function rr(x,n){return (x>>>n)|(x<<(32-n));}
 for(i=0;i<a.length;i+=64){var w=[];for(j=0;j<16;j++)w[j]=(a[i+4*j]<<24)|(a[i+4*j+1]<<16)|(a[i+4*j+2]<<8)|a[i+4*j+3];for(j=16;j<64;j++){var u=w[j-15],v=w[j-2];w[j]=(w[j-16]+(rr(u,7)^rr(u,18)^(u>>>3))+w[j-7]+(rr(v,17)^rr(v,19)^(v>>>10)))|0;}
 var x=h.slice(0);for(j=0;j<64;j++){var t=(x[7]+(rr(x[4],6)^rr(x[4],11)^rr(x[4],25))+((x[4]&x[5])^(~x[4]&x[6]))+k[j]+w[j])|0;var z=((rr(x[0],2)^rr(x[0],13)^rr(x[0],22))+((x[0]&x[1])^(x[0]&x[2])^(x[1]&x[2])))|0;x=[(t+z)|0,x[0],x[1],x[2],(x[3]+t)|0,x[4],x[5],x[6]];}for(j=0;j<8;j++)h[j]=(h[j]+x[j])|0;}
 var out='';for(i=0;i<8;i++)out+=('00000000'+(h[i]>>>0).toString(16)).slice(-8);return out;
}
function FindProxyForURL(url,host) {
 host=host.toLowerCase().replace(/\.$/,'');
 var protectedOrigin=strictProxy;for(var i=0;i<protectedSuffixes.length;i++){var d=protectedSuffixes[i].toLowerCase().replace(/\.$/,'');if(host===d||host.slice(-(d.length+1))==='.'+d)protectedOrigin=true;}
 var h=sha256(host);return 'PROXY '+h.slice(0,32)+'.'+h.slice(32)+'.localhost:%d'+(protectedOrigin?'':'; DIRECT');
}`, port)
}
func ChromeArguments(address, profile string) ([]string, error) {
	return ChromeArgumentsGuarded(address, profile, nil, false)
}
func ChromeArgumentsGuarded(address, profile string, extra []string, strict bool) ([]string, error) {
	host, p, e := net.SplitHostPort(address)
	if e != nil || host != "127.0.0.1" || profile == "" {
		return nil, fmt.Errorf("loopback transport and persistent Chrome profile required")
	}
	port, e := strconv.ParseUint(p, 10, 16)
	if e != nil || port == 0 {
		return nil, fmt.Errorf("invalid loopback port")
	}
	pac := "data:application/x-ns-proxy-autoconfig;base64," + base64.StdEncoding.EncodeToString([]byte(guardedPAC(uint16(port), extra, strict)))
	return []string{"--user-data-dir=" + profile, "--proxy-pac-url=" + pac, "--proxy-bypass-list=<-loopback>", "--no-first-run"}, nil
}
