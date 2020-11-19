package analytic

import (
    "VoipSniffer/adapter/kafka/analytic/file"
    "VoipSniffer/dao"
    "VoipSniffer/library/log"
    "VoipSniffer/prot/sip"
    "VoipSniffer/prot/udp"
	"strconv"
)

type Config struct {
	Consumer *ConsumerConfig
	Producer *ProducerConfig
}

func analyzePacket(data []byte) {
	rtd := file.Parse(string(data))

	sport, err := strconv.Atoi(rtd.Sport)
	if err != nil {
		log.Errorf("cannot convert %s to an integer", rtd.Sport)
	}
	dport, err := strconv.Atoi(rtd.Dport)
	if err != nil {
		log.Errorf("cannot convert %s to an integer", rtd.Dport)
	}

	sipMsg := sip.Parse([]byte(rtd.ParamContent))
	um := dao.UnpackedMessage{
		EventId:   rtd.EventId,
		EventTime: rtd.EventTime,
		UDP: &udp.UdpMsg{
			SrcIP:   rtd.SaddrV4,
			DstIP:   rtd.DaddrV4,
			SrcPort: uint16(sport),
			DstPort: uint16(dport),
		},
		SIP: &sipMsg,
	}
	log.Debug(data)
	dao.InsertSipPacket(&um)
}
