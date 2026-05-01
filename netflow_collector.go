package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"
)

type NetFlowV5Header struct {
	Version      uint16
	Count        uint16
	SysUptime    uint32
	UnixSecs     uint32
	UnixNsecs    uint32
	FlowSequence uint32
	EngineType   uint8
	EngineID     uint8
	Sampling     uint16
}

type NetFlowV5Record struct {
	SrcAddr  uint32
	DstAddr  uint32
	NextHop  uint32
	Input    uint16
	Output   uint16
	Packets  uint32
	Octets   uint32
	First    uint32
	Last     uint32
	SrcPort  uint16
	DstPort  uint16
	Pad1     uint8
	TCPFlags uint8
	Protocol uint8
	TOS      uint8
	SrcAS    uint16
	DstAS    uint16
	SrcMask  uint8
	DstMask  uint8
	Pad2     uint16
}

type JSONFlow struct {
	Timestamp time.Time `json:"timestamp"`
	SrcIP     string    `json:"src_ip"`
	DstIP     string    `json:"dst_ip"`
	SrcPort   uint16    `json:"src_port"`
	DstPort   uint16    `json:"dst_port"`
	Packets   uint32    `json:"packets"`
	Bytes     uint32    `json:"bytes"`
	Protocol  uint8     `json:"protocol"`
	TCPFlags  uint8     `json:"tcp_flags"`
}

func ipFromUint32(ip uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ip)
	return net.IP(b).String()
}

func startNetFlowCollector() {
	addr := ":2055"

	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer conn.Close()

	file, err := os.Create("flows.json")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	log.Printf("Listening for NetFlow v5 on %s...\n", addr)

	buffer := make([]byte, 1500)

	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		if n < 24 {
			continue
		}

		header := NetFlowV5Header{}
		headerBuf := buffer[:24]
		err = binary.Read(
			bytesReader(headerBuf),
			binary.BigEndian,
			&header,
		)
		if err != nil {
			log.Println("Header parse error:", err)
			continue
		}

		if header.Version != 5 {
			continue
		}

		offset := 24
		for i := 0; i < int(header.Count); i++ {
			if offset+48 > n {
				break
			}

			var record NetFlowV5Record
			recordBuf := buffer[offset : offset+48]

			err := binary.Read(
				bytesReader(recordBuf),
				binary.BigEndian,
				&record,
			)
			if err != nil {
				log.Println("Record parse error:", err)
				break
			}

			flow := JSONFlow{
				Timestamp: time.Now(),
				SrcIP:     ipFromUint32(record.SrcAddr),
				DstIP:     ipFromUint32(record.DstAddr),
				SrcPort:   record.SrcPort,
				DstPort:   record.DstPort,
				Packets:   record.Packets,
				Bytes:     record.Octets,
				Protocol:  record.Protocol,
				TCPFlags:  record.TCPFlags,
			}

			jsonData, err := json.Marshal(flow)
			if err != nil {
				log.Println("JSON error:", err)
				continue
			}

			file.Write(jsonData)
			file.Write([]byte("\n"))

			offset += 48
		}
	}
}

// helper to avoid importing bytes explicitly inline
func bytesReader(b []byte) *byteReader {
	return &byteReader{b: b}
}

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (int, error) {
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
