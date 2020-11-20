package udp

type UdpMsg struct {
	SrcIP   string
	DstIP   string
	SrcPort int
	DstPort int
	Payload []byte
}
