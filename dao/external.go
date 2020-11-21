package dao

import (
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"centnet-cdrrs/prot/udp"
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

type Attribution struct {
	province string
	city     string
}

func InsertSipPacket(msg *UnpackedMessage) {
	var sipPacket SipAnalyticPacket

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

//func InsertCDR(cdr VoipRestoredCdr) {
//	o := orm.NewOrm()
//	_, err := o.Insert(&cdr)
//	if err != nil {
//		log.Error(err)
//	}
//}

func GetCalleeAttribution(inviteData *SipAnalyticPacket) (province, city string) {
	sql := fmt.Sprintf("SELECT province,city FROM (SELECT to_user FROM sip_analytic_packet WHERE cseq_method = \"INVITE\" AND req_status_code = \"200\" AND call_id = %s) AS a LEFT JOIN phone_position b ON SUBSTR(a.to_user,2,7) = b.phone", inviteData.CallId)
	attribution := new(Attribution)
	err := orm.NewOrm().Raw(sql).QueryRow(&attribution)
	if err != nil {
		province = ""
		city = ""
		log.Error(err)
		return
	}

	province = attribution.province
	city = attribution.city
	return
}

func InsertCDR(cdr *VoipRestoredCdr) {
	sql := fmt.Sprintf("insert into voip_restored_cdr (%s,%s,%d,%s,%d,%s,%s,%s,%s,%s,%s,%s,%s,%d)", cdr.CallId, cdr.CallerIp, cdr.CallerPort,
		cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice, cdr.CalleeDevice, cdr.CalleeProvince,
		cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration)
	_, err := orm.NewOrm().Raw(sql).Exec()
	if err != nil {
		log.Error(err)
	}
}

func GetInvite200OKMsg(bye200ok *SipAnalyticPacket) {
	var invite200OKMsg SipAnalyticPacket
	err := orm.NewOrm().QueryTable("sip_analytic_packet").Filter("call_id", bye200ok.CallId).One(&invite200OKMsg)
	if err != nil {
		//未找到对应的200OK的包直接入库
		_, err = orm.NewOrm().Insert(bye200ok)
		log.Error(err)
	}
}
