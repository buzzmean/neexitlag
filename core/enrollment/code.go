// Package enrollment contains public enrollment wire types and pure helpers.
// Never format an activation credential in errors or diagnostic structures.
package enrollment

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/buzzmean/neexitlag/core/freshpc"
	"net"
	"strings"
)

const CodeSymbols = 24

func NewCode() (string, error) {
	b := make([]byte, 15)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	v := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	parts := []string{"VQ"}
	for i := 0; i < len(v); i += 4 {
		parts = append(parts, v[i:i+4])
	}
	return strings.Join(parts, "-"), nil
}
func CodeHash(code string) ([]byte, error) {
	if len(code) > 64 {
		return nil, errors.New("ACTIVATION_INVALID")
	}
	// Strict canonical grouping avoids aliases/leaking malformed input in errors.
	p := strings.Split(strings.ToUpper(strings.TrimSpace(code)), "-")
	if len(p) != 7 || p[0] != "VQ" {
		return nil, errors.New("ACTIVATION_INVALID")
	}
	for _, v := range p[1:] {
		if len(v) != 4 {
			return nil, errors.New("ACTIVATION_INVALID")
		}
		for _, c := range v {
			if !(c >= 'A' && c <= 'Z' || c >= '2' && c <= '7') {
				return nil, errors.New("ACTIVATION_INVALID")
			}
		}
	}
	h := sha256.Sum256([]byte(strings.Join(p, "-")))
	return h[:], nil
}

type Hardware struct {
	UUID        string `json:"uuid"`
	BoardSerial string `json:"board_serial"`
	MAC         string `json:"mac"`
}

func HardwareHash(h Hardware, salt []byte) (string, error) {
	uuid := strings.ToLower(strings.TrimSpace(h.UUID))
	board := strings.ToUpper(strings.TrimSpace(h.BoardSerial))
	compact := strings.ReplaceAll(uuid, "-", "")
	raw, e := hex.DecodeString(compact)
	if e != nil || len(raw) != 16 || compact == strings.Repeat("0", 32) || compact == strings.Repeat("f", 32) || len(board) < 2 || len(board) > 256 || len(salt) != 32 {
		return "", errors.New("HARDWARE_ID_UNAVAILABLE")
	}
	for _, bad := range []string{"DEFAULT STRING", "TO BE FILLED BY O.E.M.", "NONE", "UNKNOWN"} {
		if board == bad {
			return "", errors.New("HARDWARE_ID_UNAVAILABLE")
		}
	}
	mac, e := net.ParseMAC(strings.ReplaceAll(strings.TrimSpace(h.MAC), "-", ":"))
	if e != nil || len(mac) != 6 || mac[0]&1 != 0 || mac.String() == "00:00:00:00:00:00" {
		return "", errors.New("HARDWARE_ID_UNAVAILABLE")
	}
	b, _ := json.Marshal([]string{"veliq-machine-v2", compact, board, mac.String()})
	m := hmac.New(sha256.New, salt)
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil)), nil
}

type Request struct {
	Code   string          `json:"code"`
	Device freshpc.Request `json:"device"`
}
type Response struct {
	ProbeToken  string            `json:"probe_token"`
	DeviceID    string            `json:"device_id"`
	Certificate string            `json:"certificate"`
	Trust       map[string]string `json:"trust"`
}
type Policy struct {
	Ready          bool   `json:"enrollment_ready"`
	HardwareSalt   string `json:"hardware_salt"`
	CANotAfter     string `json:"ca_not_after"`
	ServerNotAfter string `json:"server_not_after,omitempty"`
}
