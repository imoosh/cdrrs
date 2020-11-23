package kafka

import (
	"centnet-cdrrs/adapter/kafka/file"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model"
	"centnet-cdrrs/prot/sip"
	"encoding/json"
	"strconv"
)

type Config struct {
	/* 原始数据包缓存配置 */
	SipPacketProducer *ProducerConfig

	/* 原始数据包读取配置 */
	SipPacketConsumer *ConsumerConfig

	/* 解析的数据包缓存配置 */
	RestoreCDRProducer *ProducerConfig

	/* 解析的数据包读取配置 */
	RestoreCDRConsumer *ConsumerConfig

	/* 话单数据推送配置 */
	FraudModelProducer *ProducerConfig
}

// 解析sip报文
func AnalyzePacket(consumer *Consumer, key, value interface{}) {

	rtd := file.Parse(string(value.([]byte)))
	sipMsg := sip.Parse([]byte(rtd.ParamContent))

	pkt := model.SipAnalyticPacket{
		EventId:       rtd.EventId,
		EventTime:     rtd.EventTime,
		Sip:           rtd.SaddrV4,
		Sport:         0,
		Dip:           rtd.DaddrV4,
		Dport:         0,
		CallId:        string(sipMsg.CallId.Value),
		CseqMethod:    string(sipMsg.Req.Method),
		ReqMethod:     string(sipMsg.Cseq.Method),
		ReqStatusCode: 0,
		ReqUser:       string(sipMsg.Req.User),
		ReqHost:       string(sipMsg.Req.Host),
		ReqPort:       0,
		FromName:      string(sipMsg.From.Name),
		FromUser:      string(sipMsg.From.User),
		FromHost:      string(sipMsg.From.Host),
		FromPort:      0,
		ToName:        string(sipMsg.To.Name),
		ToUser:        string(sipMsg.To.User),
		ToHost:        string(sipMsg.To.Host),
		ToPort:        0,
		ContactName:   string(sipMsg.Contact.Name),
		ContactUser:   string(sipMsg.Contact.User),
		ContactHost:   string(sipMsg.Contact.Host),
		ContactPort:   0,
		UserAgent:     string(sipMsg.Ua.Value),
	}

	var err error
	if pkt.Sport, err = strconv.Atoi(rtd.Sport); err != nil {
		log.Errorf("cannot convert %s to an integer", rtd.Sport)
		return
	}
	if pkt.Dport, err = strconv.Atoi(rtd.Dport); err != nil {
		log.Errorf("cannot convert %s to an integer", rtd.Dport)
		return
	}
	if pkt.ReqStatusCode, err = strconv.Atoi(string(sipMsg.Req.StatusCode)); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.Req.StatusCode)
		return
	}
	if pkt.ReqPort, err = strconv.Atoi(string(sipMsg.Req.Port)); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.Req.Port)
		return
	}
	if pkt.FromPort, err = strconv.Atoi(string(sipMsg.From.Port)); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.From.Port)
		return
	}
	if pkt.ToPort, err = strconv.Atoi(string(sipMsg.To.Port)); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.To.Port)
		return
	}
	if pkt.ContactPort, err = strconv.Atoi(string(sipMsg.Contact.Port)); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.Contact.Port)
		return
	}

	jsonStr, err := json.Marshal(pkt)
	if err != nil {
		log.Error(err)
		return
	}

	consumer.next.Log("", string(jsonStr))
}

func RestoreCDR(consumer *Consumer, key, value interface{}) {

	var sipMsg model.SipAnalyticPacket
	err := json.Unmarshal([]byte(value.(string)), &sipMsg)
	if err != nil {
		log.Error(err)
	}

	if sipMsg.CseqMethod == "INVITE" && sipMsg.ReqStatusCode == 200 {
		model.HandleInvite200OKMessage(key.(string), value.(string))

	} else if sipMsg.CseqMethod == "BYE" || sipMsg.ReqStatusCode == 200 {
		cdrPkt := model.HandleBye200OKMsg(key.(string), sipMsg)

		cdrStr, err := json.Marshal(&cdrPkt)
		if err != nil {
			log.Errorf("json.Marshal error: ", err)
			return
		}

		consumer.next.Log("", string(cdrStr))
		log.Debug(cdrStr)
	}
}

//func RestoreCDR() {
//	callid := `04ab01e2d142787@192.168.6.24`
//	inviteok := `
//	{
//	"id":10000,
//    "eventId":"10020044170",
//    "eventTime":"20201111100820",
//    "sip":"219.143.187.139",
//    "sport":5088,
//    "dip":"192.168.6.24",
//    "dport":5060,
//    "callId":"04ab01e2d142787@192.168.6.24",
//    "cseqMethod":"INVITE",
//    "reqMethod":" ",
//    "reqStatusCode":200,
//    "reqUser":"018926798345",
//    "reqHost":"219.143.187.139",
//    "reqPort":5060,
//    "fromName":"1101385",
//    "fromUser":"1101385",
//    "fromHost":"192.168.6.24",
//    "fromPort":5060,
//    "toName":" ",
//    "toUser":"018926798345",
//    "toHost":"219.143.187.139",
//    "toPort":5060,
//    "contactName":"1101385",
//    "contactUser":"1101385",
//    "contactHost":"192.168.6.24",
//    "contactPort":5060,
//    "userAgent":"DonJin SIP Server 3.2.0_i"
//	}
//	`
//	cdr.ParseInvite200OKMessage(callid, inviteok)
//	call := `04ab01e2d142787@192.168.6.24`
//	byeOk := `
//	{
//	"id":10000,
//	"eventId":"10020044170",
//	"eventTime":"20201111100903",
//	"sip":"219.143.187.139",
//	"sport":5088,
//	"dip":"192.168.6.24",
//	"dport":5060,
//	"callId":"04ab01e2d142787@192.168.6.24",
//	"cseqMethod":"BYE",
//	"reqMethod":"100",
//	"reqStatusCode":200,
//	"reqUser":"",
//	"reqHost":"",
//	"reqPort":5060,
//	"fromName":"1101385",
//	"fromUser":"1101385",
//	"fromHost":"192.168.6.24",
//	"fromPort":5060,
//	"toName":"",
//	"toUser":"018926798345",
//	"toHost":"219.143.187.139",
//	"toPort":"5060",
//	"contactName":"1101385",
//	"contactUser":"1101385",
//	"contactHost":"192.168.6.24",
//	"contactPort":"5060",
//	"userAgent":""
//	}`
//	//time.Sleep(time.Duration(6) * time.Second)
//	cdr.ParseBye200OKMsg(call, byeOk)
//}
