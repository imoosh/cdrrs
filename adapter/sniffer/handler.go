package sniffer

import (
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"centnet-cdrrs/prot/udp"
	"encoding/json"
	"fmt"
)

type rawData struct {
	SrcIP   string
	DstIP   string
	SrcPort int
	DstPort int
	Payload []byte
}

func (data *rawData) String() string {
	return fmt.Sprintf("%s:%d - %s:%d %s", data.SrcIP, data.SrcPort, data.DstIP, data.DstPort, data.Payload)
}

func doPacket(ps *PacketSniffer, data *rawData) {
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

	jsonStr, err := json.Marshal(um)
	if err != nil {
		log.Error(err)
		return
	}
	ps.producer.Log("", string(jsonStr))
}
