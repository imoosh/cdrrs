package sniffer

import (
	"bytes"
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/library/log"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
//device            = "en0"
//snapLen     int32 = 65535
//filter            = "udp"
//packetCount       = 0
)

type Config struct {
	handle  *pcap.Handle
	Device  string
	SnapLen int32
	Filter  string
}

type PacketSniffer struct {
	config   *Config
	producer *kafka.Producer
}

func NewPacketSniffer(c *Config, producer *kafka.Producer) *PacketSniffer {
	return &PacketSniffer{
		config:   c,
		producer: producer,
	}
}

func (ps *PacketSniffer) CustomFilter(ipLayer *layers.IPv4, udpLayer *layers.UDP) bool {
	//fmt.Println(ipLayer.SrcIP, ipLayer.DstIP, uint16(udpLayer.SrcPort), uint16(udpLayer.DstPort))

	if !bytes.HasPrefix(udpLayer.Payload, []byte("INVITE sip:")) &&
		!bytes.HasPrefix(udpLayer.Payload, []byte("PRACK sip:")) &&
		!bytes.HasPrefix(udpLayer.Payload, []byte("ACK sip:")) &&
		!bytes.HasPrefix(udpLayer.Payload, []byte("BYE sip:")) &&
		!bytes.HasPrefix(udpLayer.Payload, []byte("CANCEL sip:")) &&
		!bytes.HasPrefix(udpLayer.Payload, []byte("REGISTER sip:")) &&
		!bytes.HasPrefix(udpLayer.Payload, []byte("SIP/2.0")) {
		return false
	}

	return true
}

func (ps *PacketSniffer) Run() error {
	handle, err := pcap.OpenLive(ps.config.Device, ps.config.SnapLen, true, pcap.BlockForever)
	if err != nil {
		log.Error(err)
		return errors.New(fmt.Sprintf("OpenLive '%s' failed", ps.config.Device))
	}

	if err = handle.SetBPFFilter(ps.config.Filter); err != nil {
		log.Error(err)
		return errors.New(fmt.Sprintf("SetBPFFilter '%s' failed", ps.config.Filter))
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true
	for packet := range packetSource.Packets() {
		if packet.NetworkLayer() == nil || packet.NetworkLayer().LayerType() != layers.LayerTypeIPv4 ||
			packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeUDP {
			continue
		}

		ipLayer := packet.NetworkLayer().(*layers.IPv4)
		udpLayer := packet.TransportLayer().(*layers.UDP)
		if !ps.CustomFilter(ipLayer, udpLayer) {
			continue
		}

		data := &rawData{
			SrcIP:   ipLayer.SrcIP.String(),
			DstIP:   ipLayer.DstIP.String(),
			SrcPort: int(udpLayer.SrcPort),
			DstPort: int(udpLayer.DstPort),
			Payload: udpLayer.Payload,
		}

		doPacket(ps, data)
	}

	return nil
}
