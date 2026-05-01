package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

//go:embed flow.o
var bpfProgram []byte

type FlowEvent struct {
	SrcMAC [6]byte
	DstMAC [6]byte

	SrcIP uint32
	DstIP uint32

	SrcPort uint16
	DstPort uint16

	Proto    uint8
	ICMPType uint8
	ICMPCode uint8
	_        [1]byte

	Bytes uint32

	PayloadLen uint16
	Payload    [256]byte
}

type DNSInfo struct {
	QueryName string
	Answers   []string
}

func main() {
	log.Println("Starting flow agent...")

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(bpfProgram))
	if err != nil {
		log.Fatal(err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		log.Fatal(err)
	}
	defer coll.Close()

	prog := coll.Programs["flow_monitor"]
	events := coll.Maps["events"]

	iface, err := net.InterfaceByName(getInterface())
	if err != nil {
		log.Fatal(err)
	}

	l, err := link.AttachTCX(link.TCXOptions{
		Program:   prog,
		Interface: iface.Index,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	rd, err := ringbuf.NewReader(events)
	if err != nil {
		log.Fatal(err)
	}
	defer rd.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Flow agent running...")

	for {
		select {
		case <-stop:
			return

		default:
			record, err := rd.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					return
				}
				continue
			}

			if len(record.RawSample) < int(unsafe.Sizeof(FlowEvent{})) {
				continue
			}

			evt := *(*FlowEvent)(unsafe.Pointer(&record.RawSample[0]))
			logJSON(evt)
		}
	}
}

func getInterface() string {
	if v := os.Getenv("INTERFACE"); v != "" {
		return v
	}
	return "fmdefaultvpc"
}

func ipToString(ip uint32) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], ip)
	return net.IP(b[:]).String()
}

func macToString(mac [6]byte) string {
	return net.HardwareAddr(mac[:]).String()
}

func protocolToString(p uint8) string {
	switch p {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 1:
		return "ICMP"
	case 0:
		return "ARP"
	default:
		return "OTHER"
	}
}

func logJSON(evt FlowEvent) {
	payload := evt.Payload[:evt.PayloadLen]
	var http HTTPInfo
	var sni string

	// HTTP only on port 80
	if evt.Proto == 6 && (evt.SrcPort == 80 || evt.DstPort == 80) {
		http = parseHTTP(payload)
	}

	// TLS only on port 443
	if evt.Proto == 6 && (evt.SrcPort == 443 || evt.DstPort == 443) {
		sni = parseTLSClientHello(payload)
	}

	out := map[string]interface{}{
		"src_mac":  macToString(evt.SrcMAC),
		"dst_mac":  macToString(evt.DstMAC),
		"src_ip":   ipToString(evt.SrcIP),
		"dst_ip":   ipToString(evt.DstIP),
		"protocol": protocolToString(evt.Proto),
		"bytes":    evt.Bytes,
	}

	if evt.SrcPort != 0 {
		out["src_port"] = evt.SrcPort
	}
	if evt.DstPort != 0 {
		out["dst_port"] = evt.DstPort
	}

	// HTTP
	if http.Method != "" {
		out["http_method"] = http.Method
		out["http_host"] = http.Host
		out["http_path"] = http.Path
	}

	// 🔥 TLS (HTTPS visibility)
	if sni != "" {
		out["tls_sni"] = sni
	}

	// ✅ DNS parsing (only UDP/53)
	if evt.Proto == 17 && (evt.SrcPort == 53 || evt.DstPort == 53) {
		dns := parseDNS(evt.Payload[:evt.PayloadLen])
		if dns.QueryName != "" {
			out["dns_query"] = dns.QueryName
		}
		if len(dns.Answers) > 0 {
			out["dns_answers"] = dns.Answers
		}
	}

	data, _ := json.Marshal(out)
	os.Stdout.Write(data)
	os.Stdout.Write([]byte("\n"))
}

//
// ===== DNS PARSER =====
//

func parseDNS(p []byte) DNSInfo {
	var d DNSInfo
	if len(p) < 12 {
		return d
	}

	qd := int(binary.BigEndian.Uint16(p[4:6]))
	an := int(binary.BigEndian.Uint16(p[6:8]))

	off := 12

	if qd > 0 {
		name, o, ok := readName(p, off)
		if !ok {
			return d
		}
		d.QueryName = name
		off = o + 4
	}

	for i := 0; i < an; i++ {
		_, o, ok := readName(p, off)
		if !ok {
			break
		}
		off = o

		if off+10 > len(p) {
			break
		}

		t := binary.BigEndian.Uint16(p[off : off+2])
		l := int(binary.BigEndian.Uint16(p[off+8 : off+10]))
		off += 10

		if off+l > len(p) {
			break
		}

		if t == 1 && l == 4 {
			d.Answers = append(d.Answers, net.IP(p[off:off+4]).String())
		}

		off += l
	}

	return d
}

func readName(msg []byte, off int) (string, int, bool) {
	var parts []string

	for {
		if off >= len(msg) {
			return "", 0, false
		}

		l := int(msg[off])

		if l == 0 {
			off++
			break
		}

		if l&0xC0 == 0xC0 {
			if off+1 >= len(msg) {
				return "", 0, false
			}
			ptr := int(binary.BigEndian.Uint16(msg[off:off+2]) & 0x3FFF)
			name, _, ok := readName(msg, ptr)
			if !ok {
				return "", 0, false
			}
			parts = append(parts, name)
			off += 2
			break
		}

		off++
		if off+l > len(msg) {
			return "", 0, false
		}

		parts = append(parts, string(msg[off:off+l]))
		off += l
	}

	return strings.Join(parts, "."), off, true
}

func parseTLSClientHello(data []byte) string {
	// Need at least TLS record header
	if len(data) < 5 {
		return ""
	}

	// TLS Handshake (0x16)
	if data[0] != 0x16 {
		return ""
	}

	// Record length
	recordLen := int(binary.BigEndian.Uint16(data[3:5]))
	if recordLen+5 > len(data) {
		return ""
	}

	// Handshake type (ClientHello = 0x01)
	if data[5] != 0x01 {
		return ""
	}

	i := 5

	// Skip handshake header (1 type + 3 length)
	i += 4

	// Skip version (2) + random (32)
	i += 34
	if i >= len(data) {
		return ""
	}

	// Session ID
	sidLen := int(data[i])
	i += 1 + sidLen
	if i >= len(data) {
		return ""
	}

	// Cipher suites
	if i+2 > len(data) {
		return ""
	}
	csLen := int(binary.BigEndian.Uint16(data[i:]))
	i += 2 + csLen
	if i >= len(data) {
		return ""
	}

	// Compression methods
	compLen := int(data[i])
	i += 1 + compLen
	if i >= len(data) {
		return ""
	}

	// Extensions
	if i+2 > len(data) {
		return ""
	}
	extLen := int(binary.BigEndian.Uint16(data[i:]))
	i += 2

	end := i + extLen

	for i+4 <= end && i+4 <= len(data) {
		extType := binary.BigEndian.Uint16(data[i:])
		extSize := int(binary.BigEndian.Uint16(data[i+2:]))

		i += 4

		if extType == 0x0000 { // SNI
			if i+5 > len(data) {
				return ""
			}

			// skip list len (2) + name type (1)
			nameLen := int(binary.BigEndian.Uint16(data[i+3:]))

			if i+5+nameLen > len(data) {
				return ""
			}

			return string(data[i+5 : i+5+nameLen])
		}

		i += extSize
	}

	return ""
}

type HTTPInfo struct {
	Method string
	Host   string
	Path   string
}

func parseHTTP(p []byte) HTTPInfo {
	var h HTTPInfo

	// Quick sanity check
	if len(p) < 16 {
		return h
	}

	// Convert once
	s := string(p)

	// 🔒 Strict method detection (avoids TLS/DNS misclassification)
	if !(strings.HasPrefix(s, "GET ") ||
		strings.HasPrefix(s, "POST ") ||
		strings.HasPrefix(s, "PUT ") ||
		strings.HasPrefix(s, "DELETE ") ||
		strings.HasPrefix(s, "HEAD ") ||
		strings.HasPrefix(s, "OPTIONS ") ||
		strings.HasPrefix(s, "PATCH ")) {
		return h
	}

	// Split headers
	lines := strings.Split(s, "\r\n")
	if len(lines) == 0 {
		return h
	}

	// ---- Request line ----
	// Example: GET /path HTTP/1.1
	parts := strings.Split(lines[0], " ")
	if len(parts) >= 2 {
		h.Method = parts[0]
		h.Path = parts[1]
	}

	// ---- Headers ----
	for _, line := range lines[1:] {
		if line == "" {
			break // end of headers
		}

		// Case-insensitive match
		if len(line) >= 5 && strings.EqualFold(line[:5], "Host:") {
			h.Host = strings.TrimSpace(line[5:])
			break
		}
	}

	return h
}
