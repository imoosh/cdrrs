package dao

import (
    "VoipSniffer/prot/sip"
    "VoipSniffer/prot/udp"
	"fmt"
	"github.com/astaxie/beego/orm"
	"strconv"
)

type UnpackedMessage struct {
	EventId   string
	EventTime string
	UDP       *udp.UdpMsg
	SIP       *sip.SipMsg
}

func InsertSipPacket(msg *UnpackedMessage) {
	var sipPacket Sip

	sipPacket.EventId = msg.EventId
	sipPacket.EventTime = msg.EventTime

	sipPacket.Sip = msg.UDP.SrcIP
	sipPacket.Sport = msg.UDP.SrcPort
	sipPacket.Dip = msg.UDP.DstIP
	sipPacket.Dport = msg.UDP.DstPort

	sipPacket.CallId = string(msg.SIP.CallId.Value)
	sipPacket.ReqMethod = string(msg.SIP.Req.Method)
	sipPacket.ReqStatusCode, _ = strconv.Atoi(string(msg.SIP.Req.StatusCode))
	sipPacket.ReqUser = string(msg.SIP.Req.User)
	sipPacket.ReqHost = string(msg.SIP.Req.Host)
	sipPacket.ReqPort, _ = strconv.Atoi(string(msg.SIP.Req.Port))
	if sipPacket.ReqPort == 0 {
		sipPacket.ReqPort = 5060
	}
	sipPacket.FromName = string(msg.SIP.From.Name)
	sipPacket.FromUser = string(msg.SIP.From.User)
	sipPacket.FromHost = string(msg.SIP.From.Host)
	sipPacket.FromPort, _ = strconv.Atoi(string(msg.SIP.From.Port))
	if sipPacket.FromPort == 0 {
		sipPacket.FromPort = 5060
	}
	sipPacket.ToName = string(msg.SIP.To.Name)
	sipPacket.ToUser = string(msg.SIP.To.User)
	sipPacket.ToHost = string(msg.SIP.To.Host)
	sipPacket.ToPort, _ = strconv.Atoi(string(msg.SIP.To.Port))
	if sipPacket.ToPort == 0 {
		sipPacket.ToPort = 5060
	}
	sipPacket.ContactName = string(msg.SIP.Contact.Name)
	sipPacket.ContactUser = string(msg.SIP.Contact.User)
	sipPacket.ContactHost = string(msg.SIP.Contact.Host)
	sipPacket.ContactPort, _ = strconv.Atoi(string(msg.SIP.Contact.Port))
	if sipPacket.ContactPort == 0 {
		sipPacket.ContactPort = 5060
	}
	sipPacket.CseqMethod = string(msg.SIP.Cseq.Method)
	sipPacket.UserAgent = string(msg.SIP.Ua.Value)

	o := orm.NewOrm()
	_, err := o.Insert(&sipPacket)
	if err != nil {
		fmt.Println(err)
	}
}
