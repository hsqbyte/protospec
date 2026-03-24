package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
)

type pcapOptions struct {
	Protocol string // -p filter
	Filter   string // --filter "field==value"
	Format   string // --format json|table|summary
}

func runPcap(ctx *Context, args []string) error {
	opts := pcapOptions{Format: "summary"}
	var file string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p":
			i++
			if i < len(args) {
				opts.Protocol = strings.ToUpper(args[i])
			}
		case "--filter":
			i++
			if i < len(args) {
				opts.Filter = args[i]
			}
		case "--format":
			i++
			if i < len(args) {
				opts.Format = args[i]
			}
		default:
			file = args[i]
		}
	}

	if file == "" {
		return fmt.Errorf("usage: psl pcap [options] <file.pcap>")
	}

	return decodePcapFile(ctx, file, opts)
}

func decodePcapFile(ctx *Context, file string, opts pcapOptions) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	reader, err := openPcapReader(f)
	if err != nil {
		return fmt.Errorf("failed to open pcap: %w", err)
	}

	packetSource := gopacket.NewPacketSource(reader, reader.LinkType())
	packetNum := 0

	for packet := range packetSource.Packets() {
		packetNum++
		result := decodePacketLayers(ctx, packet)

		// Protocol filter
		if opts.Protocol != "" {
			if !resultContainsProtocol(result, opts.Protocol) {
				continue
			}
		}

		// Field filter
		if opts.Filter != "" {
			if !matchFieldFilter(result, opts.Filter) {
				continue
			}
		}

		switch opts.Format {
		case "json":
			printPacketJSON(packetNum, result)
		case "table":
			printPacketTable(packetNum, result)
		default:
			printPacketSummary(packetNum, packet, result)
		}
	}

	return nil
}

// unifiedReader wraps pcapgo readers to provide a unified interface.
type unifiedReader struct {
	r        gopacket.PacketDataSource
	linkType layers.LinkType
}

func (pr *unifiedReader) LinkType() layers.LinkType { return pr.linkType }
func (pr *unifiedReader) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	return pr.r.ReadPacketData()
}

func openPcapReader(r io.Reader) (*unifiedReader, error) {
	// Try pcapng first
	ngReader, err := pcapgo.NewNgReader(r, pcapgo.DefaultNgReaderOptions)
	if err == nil {
		return &unifiedReader{r: ngReader, linkType: ngReader.LinkType()}, nil
	}

	// Seek back and try pcap
	if seeker, ok := r.(io.ReadSeeker); ok {
		seeker.Seek(0, io.SeekStart)
	}
	pcReader, err := pcapgo.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a valid pcap/pcapng file")
	}
	return &unifiedReader{r: pcReader, linkType: pcReader.LinkType()}, nil
}

// layerResult holds decoded data for one protocol layer.
type layerResult struct {
	Protocol string         `json:"protocol"`
	Fields   map[string]any `json:"fields"`
}

func decodePacketLayers(ctx *Context, packet gopacket.Packet) []layerResult {
	var results []layerResult

	for _, layer := range packet.Layers() {
		lr := layerResult{
			Protocol: layer.LayerType().String(),
			Fields:   make(map[string]any),
		}

		switch l := layer.(type) {
		case *layers.Ethernet:
			lr.Protocol = "Ethernet"
			lr.Fields["srcMAC"] = l.SrcMAC.String()
			lr.Fields["dstMAC"] = l.DstMAC.String()
			lr.Fields["etherType"] = fmt.Sprintf("0x%04x", uint16(l.EthernetType))
		case *layers.IPv4:
			lr.Protocol = "IPv4"
			lr.Fields["srcIP"] = l.SrcIP.String()
			lr.Fields["dstIP"] = l.DstIP.String()
			lr.Fields["version"] = l.Version
			lr.Fields["ihl"] = l.IHL
			lr.Fields["ttl"] = l.TTL
			lr.Fields["protocol"] = l.Protocol
			lr.Fields["totalLength"] = l.Length
		case *layers.IPv6:
			lr.Protocol = "IPv6"
			lr.Fields["srcIP"] = l.SrcIP.String()
			lr.Fields["dstIP"] = l.DstIP.String()
			lr.Fields["nextHeader"] = l.NextHeader
			lr.Fields["hopLimit"] = l.HopLimit
		case *layers.TCP:
			lr.Protocol = "TCP"
			lr.Fields["srcPort"] = uint16(l.SrcPort)
			lr.Fields["dstPort"] = uint16(l.DstPort)
			lr.Fields["seq"] = l.Seq
			lr.Fields["ack"] = l.Ack
			lr.Fields["flags"] = tcpFlags(l)
			lr.Fields["window"] = l.Window
		case *layers.UDP:
			lr.Protocol = "UDP"
			lr.Fields["srcPort"] = uint16(l.SrcPort)
			lr.Fields["dstPort"] = uint16(l.DstPort)
			lr.Fields["length"] = l.Length
		case *layers.DNS:
			lr.Protocol = "DNS"
			lr.Fields["id"] = l.ID
			lr.Fields["qr"] = l.QR
			lr.Fields["opcode"] = l.OpCode
			lr.Fields["qdCount"] = l.QDCount
			lr.Fields["anCount"] = l.ANCount
			if len(l.Questions) > 0 {
				lr.Fields["query"] = string(l.Questions[0].Name)
				lr.Fields["queryType"] = l.Questions[0].Type.String()
			}
		case *layers.ARP:
			lr.Protocol = "ARP"
			lr.Fields["operation"] = l.Operation
			lr.Fields["srcMAC"] = fmt.Sprintf("%x", l.SourceHwAddress)
			lr.Fields["srcIP"] = fmt.Sprintf("%d.%d.%d.%d", l.SourceProtAddress[0], l.SourceProtAddress[1], l.SourceProtAddress[2], l.SourceProtAddress[3])
		case *layers.ICMPv4:
			lr.Protocol = "ICMP"
			lr.Fields["type"] = l.TypeCode.Type()
			lr.Fields["code"] = l.TypeCode.Code()
			lr.Fields["id"] = l.Id
			lr.Fields["seq"] = l.Seq
		default:
			// Use gopacket's layer type name
			lr.Fields["raw_length"] = len(layer.LayerContents())
		}

		results = append(results, lr)
	}

	// Also try PSL-based decoding for layers we know about
	tryPSLDecode(ctx, packet, &results)

	return results
}

// tryPSLDecode attempts to decode raw layer data using PSL definitions.
func tryPSLDecode(ctx *Context, packet gopacket.Packet, results *[]layerResult) {
	// For application layer payload, try to decode with registered protocols
	if app := packet.ApplicationLayer(); app != nil {
		payload := app.Payload()
		if len(payload) > 0 {
			// Try HTTP detection (simple heuristic)
			if isHTTPPayload(payload) {
				lr := layerResult{Protocol: "HTTP", Fields: make(map[string]any)}
				lines := strings.SplitN(string(payload), "\r\n", 2)
				if len(lines) > 0 {
					lr.Fields["firstLine"] = lines[0]
				}
				lr.Fields["length"] = len(payload)
				*results = append(*results, lr)
			}
		}
	}
}

func isHTTPPayload(data []byte) bool {
	s := string(data)
	return strings.HasPrefix(s, "GET ") || strings.HasPrefix(s, "POST ") ||
		strings.HasPrefix(s, "PUT ") || strings.HasPrefix(s, "DELETE ") ||
		strings.HasPrefix(s, "HTTP/") || strings.HasPrefix(s, "HEAD ") ||
		strings.HasPrefix(s, "OPTIONS ") || strings.HasPrefix(s, "PATCH ")
}

func tcpFlags(t *layers.TCP) string {
	var flags []string
	if t.SYN {
		flags = append(flags, "SYN")
	}
	if t.ACK {
		flags = append(flags, "ACK")
	}
	if t.FIN {
		flags = append(flags, "FIN")
	}
	if t.RST {
		flags = append(flags, "RST")
	}
	if t.PSH {
		flags = append(flags, "PSH")
	}
	if t.URG {
		flags = append(flags, "URG")
	}
	return strings.Join(flags, ",")
}

func resultContainsProtocol(results []layerResult, proto string) bool {
	for _, r := range results {
		if strings.EqualFold(r.Protocol, proto) {
			return true
		}
	}
	return false
}

func matchFieldFilter(results []layerResult, filter string) bool {
	// Simple "field==value" filter
	parts := strings.SplitN(filter, "==", 2)
	if len(parts) != 2 {
		return true // invalid filter, pass through
	}
	fieldName := strings.TrimSpace(parts[0])
	fieldValue := strings.TrimSpace(parts[1])

	for _, r := range results {
		if v, ok := r.Fields[fieldName]; ok {
			if fmt.Sprintf("%v", v) == fieldValue {
				return true
			}
		}
	}
	return false
}

func printPacketJSON(num int, results []layerResult) {
	out := map[string]any{
		"packet": num,
		"layers": results,
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}

func printPacketTable(num int, results []layerResult) {
	fmt.Printf("%s#%d%s", cBold, num, cReset)
	for _, r := range results {
		fmt.Printf("  %s%s%s", cCyan, r.Protocol, cReset)
	}
	fmt.Println()
	for _, r := range results {
		for k, v := range r.Fields {
			fmt.Printf("  %s%s%s=%v", cGreen, k, cReset, v)
		}
	}
	fmt.Println()
}

func printPacketSummary(num int, packet gopacket.Packet, results []layerResult) {
	// Build protocol chain
	var protos []string
	for _, r := range results {
		protos = append(protos, r.Protocol)
	}
	chain := strings.Join(protos, " → ")

	// Summary line
	ts := packet.Metadata().Timestamp.Format("15:04:05.000000")
	length := packet.Metadata().Length

	summary := buildSummaryInfo(results)

	fmt.Printf("%s%4d%s %s %s%4d%s %s%s%s %s\n",
		cDim, num, cReset,
		ts,
		cYellow, length, cReset,
		cCyan, chain, cReset,
		summary)
}

func buildSummaryInfo(results []layerResult) string {
	var parts []string
	for _, r := range results {
		switch r.Protocol {
		case "IPv4", "IPv6":
			if src, ok := r.Fields["srcIP"]; ok {
				if dst, ok := r.Fields["dstIP"]; ok {
					parts = append(parts, fmt.Sprintf("%v → %v", src, dst))
				}
			}
		case "TCP", "UDP":
			if sp, ok := r.Fields["srcPort"]; ok {
				if dp, ok := r.Fields["dstPort"]; ok {
					parts = append(parts, fmt.Sprintf(":%v → :%v", sp, dp))
				}
			}
			if flags, ok := r.Fields["flags"]; ok && flags != "" {
				parts = append(parts, fmt.Sprintf("[%v]", flags))
			}
		case "DNS":
			if q, ok := r.Fields["query"]; ok {
				parts = append(parts, fmt.Sprintf("query=%v", q))
			}
		}
	}
	return strings.Join(parts, " ")
}
