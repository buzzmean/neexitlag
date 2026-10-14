package localstrategy

import (
	"encoding/binary"
	"strings"
)

// Inspection is passive metadata from the first eligible payload, never retained
// packet bytes. A reason is not evidence that a packet was successfully sent.
type Inspection struct {
	ObservedSNI              string `json:"observed_sni,omitempty"`
	SNIStatus                string `json:"sni_status,omitempty"`
	SegmentsSeen             int    `json:"segments_seen,omitempty"`
	SegmentsBuffered         int    `json:"segments_buffered,omitempty"`
	BytesBuffered            int    `json:"bytes_buffered,omitempty"`
	ExpectedClientHelloBytes int    `json:"expected_clienthello_bytes,omitempty"`
	ReassemblyMS             int64  `json:"reassembly_ms"`
	OutOfOrder               bool   `json:"out_of_order,omitempty"`
	ReassembledSNI           bool   `json:"reassembled_sni,omitempty"`
	TransformApplied         bool   `json:"transform_applied"`

	OutputPacketsPlanned int  `json:"output_packets_planned,omitempty"`
	OutputPacketsSent    int  `json:"output_packets_sent,omitempty"`
	SendFailed           bool `json:"send_failed,omitempty"`

	Reason         string `json:"reason"`
	Detail         string `json:"detail,omitempty"`
	Domain         string `json:"domain,omitempty"`
	PayloadBytes   int    `json:"payload_bytes"`
	RecordBytes    int    `json:"record_bytes,omitempty"`
	HandshakeBytes int    `json:"handshake_bytes,omitempty"`
}

func inspectHello(p []byte) *Inspection {
	d := &Inspection{Reason: "NO_TRANSFORM", Detail: "not a recognized complete ClientHello", PayloadBytes: len(p)}
	if len(p) == 0 || len(p) > MaxHelloBytes || p[0] != 22 {
		return d
	}
	if len(p) < 5 {
		d.Reason = "CLIENTHELLO_INCOMPLETE"
		d.Detail = "TLS record header incomplete"
		return d
	}
	if p[1] != 3 || p[2] < 1 || p[2] > 3 {
		return d
	}
	d.RecordBytes = 5 + int(binary.BigEndian.Uint16(p[3:5]))
	if d.RecordBytes > len(p) {
		d.Reason = "CLIENTHELLO_INCOMPLETE"
		d.Detail = "declared TLS record exceeds observed payload; no reassembly"
		return d
	}
	if d.RecordBytes <= 5 || p[5] != 1 {
		return d
	}
	if d.RecordBytes < 9 {
		d.Reason = "CLIENTHELLO_INCOMPLETE"
		d.Detail = "ClientHello header spans records"
		return d
	}
	d.HandshakeBytes = 9 + (int(p[6])<<16 | int(p[7])<<8 | int(p[8]))
	if d.HandshakeBytes > d.RecordBytes {
		d.Reason = "CLIENTHELLO_INCOMPLETE"
		d.Detail = "ClientHello spans TLS records; no reassembly"
		return d
	}
	if start, n, ok := tlsSNI(p); ok {
		d.Domain = strings.Clone(string(p[start : start+n]))
		d.Detail = "complete ClientHello with observed outer SNI"
		return d
	}
	// Only report missing SNI for a structurally valid complete ClientHello.
	end := d.RecordBytes
	if end < 44 || d.HandshakeBytes != end || p[9] != 3 || p[10] < 1 || p[10] > 3 || p[43] > 32 {
		return d
	}
	i := 44 + int(p[43])
	if i+2 > end {
		return d
	}
	n := int(binary.BigEndian.Uint16(p[i:]))
	i += 2
	if n < 2 || n%2 != 0 || i+n+1 > end {
		return d
	}
	i += n
	n = int(p[i])
	i++
	if n < 1 || i+n > end {
		return d
	}
	i += n
	if i == end {
		d.Reason = "SNI_NOT_FOUND"
		d.Detail = "complete ClientHello without extensions"
		return d
	}
	if i+2 > end {
		return d
	}
	n = int(binary.BigEndian.Uint16(p[i:]))
	i += 2
	if i+n != end {
		return d
	}
	for i < end {
		if i+4 > end {
			return d
		}
		kind := binary.BigEndian.Uint16(p[i:])
		n = int(binary.BigEndian.Uint16(p[i+2:]))
		i += 4
		if i+n > end || kind == 0 {
			return d
		}
		i += n
	}
	d.Reason = "SNI_NOT_FOUND"
	d.Detail = "complete ClientHello has no server_name extension"
	return d
}
