package udp

type UdpMsg struct {
    SrcIP   string
    DstIP   string
    SrcPort uint16
    DstPort uint16
    Payload []byte
}
