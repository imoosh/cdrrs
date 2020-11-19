package sniffer

import (
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"centnet-cdrrs/prot/udp"
	"fmt"
)

type rawData struct {
	SrcIP   string
	DstIP   string
	SrcPort uint16
	DstPort uint16
	Payload []byte
}

func (data *rawData) String() string {
	return fmt.Sprintf("%s:%d - %s:%d %s", data.SrcIP, data.SrcPort, data.DstIP, data.DstPort, data.Payload)
}

func doPacket(data *rawData) {
	sipMsg := sip.Parse(data.Payload)
	um := dao.UnpackedMessage{
		UDP: &udp.UdpMsg{
			SrcIP:   data.SrcIP,
			DstIP:   data.DstIP,
			SrcPort: data.SrcPort,
			DstPort: data.DstPort,
		},
		SIP: &sipMsg,
	}
	log.Debug(data)
	dao.InsertSipPacket(&um)
}
