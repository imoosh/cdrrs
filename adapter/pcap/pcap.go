package pcap

import (
    "VoipSniffer/dao"
    "VoipSniffer/prot/sip"
    "VoipSniffer/prot/udp"
    "bytes"
    //"VoipSniffer/dao"
    //"VoipSniffer/prot/sip"
    //"bytes"
    "errors"
    "fmt"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/pcap"
)

var (
    devName           = "en0"
    snapLen     int32 = 65535
    filter            = "udp"
    packetCount       = 0
)

func StartCapture() error {
    handle, err := pcap.OpenLive(devName, snapLen, true, pcap.BlockForever)
    if err != nil {
        fmt.Println(err)
        return errors.New(fmt.Sprintf("OpenLive '%s' failed", devName))
    }

    if err = handle.SetBPFFilter(filter); err != nil {
        fmt.Println(err)
        return errors.New(fmt.Sprintf("SetBPFFilter '%s' failed", filter))
    }

    ps := gopacket.NewPacketSource(handle, handle.LinkType())
    ps.NoCopy = true
    for packet := range ps.Packets() {
        if packet.NetworkLayer() == nil || packet.NetworkLayer().LayerType() != layers.LayerTypeIPv4 ||
            packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeUDP {
            continue
        }

        ipLayer := packet.NetworkLayer().(*layers.IPv4)
        udpLayer := packet.TransportLayer().(*layers.UDP)
        //fmt.Println(ipLayer.SrcIP, ipLayer.DstIP, uint16(udpLayer.SrcPort), uint16(udpLayer.DstPort))

        if !bytes.HasPrefix(udpLayer.Payload, []byte("INVITE sip:")) &&
            !bytes.HasPrefix(udpLayer.Payload, []byte("PRACK sip:")) &&
            !bytes.HasPrefix(udpLayer.Payload, []byte("ACK sip:")) &&
            !bytes.HasPrefix(udpLayer.Payload, []byte("BYE sip:")) &&
            !bytes.HasPrefix(udpLayer.Payload, []byte("CANCEL sip:")) &&
            !bytes.HasPrefix(udpLayer.Payload, []byte("REGISTER sip:")) &&
            !bytes.HasPrefix(udpLayer.Payload, []byte("SIP/2.0")) {
            continue
        }

        //fmt.Println(string(udp.Payload))
        packetCount ++
        //fmt.Println(packetCount)

        sipMsg := sip.Parse(udpLayer.Payload)
        um := dao.UnpackedMessage{
            UDP: &udp.UdpMsg{
                SrcIP:    ipLayer.SrcIP.String(),
                DstIP:    ipLayer.DstIP.String(),
                SrcPort:  uint16(udpLayer.SrcPort),
                DstPort:  uint16(udpLayer.DstPort),
            },
            SIP: &sipMsg,
        }

        dao.InsertSipPacket(&um)
    }

    return nil
}
