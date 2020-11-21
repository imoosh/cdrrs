package kafka

import (
	"centnet-cdrrs/adapter/kafka/file"
	"centnet-cdrrs/cdr"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
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

	pkt := dao.SipAnalyticPacket{
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

//func RestoreCDR(consumer *Consumer, key, value interface{}) {
//	cdr.ParseInvite200OKMessage(key.([]byte), value.([]byte))
//	cdrPkt := cdr.ParseBye200OKMsg(key.([]byte), value.([]byte))
//	jsonStr, err := json.Marshal(cdrPkt)
//	if err != nil {
//		log.Error(err)
//		return
//	}
//
//	log.Debug(jsonStr)
//	consumer.next.Log("", string(jsonStr))
//}

func RestoreCDR(consumer *Consumer, key, value interface{}) {

	callid := "04ab01e2d142787@192.168.6.24"
	inviteOk := []byte(`
	{
		"event_id": "10020044170",
  		"event_time": "20201111100820",
  		"sip": "219.143.187.139",
  		"sport": "5088",
  		"dip": "192.168.6.24",
  		"dport": "5060",
  		"call_id": "04ab01e2d142787@192.168.6.24",
	    "cseq_method": "INVITE",
	    "req_method": "",
	    "req_status_code": "200",
	    "req_user": "018926798345",
	    "req_host": "219.143.187.139",
	    "req_port": "5060",
	    "from_name": "1101385",
	    "from_user": "1101385",
	    "from_host": "192.168.6.24",
	    "from_port": "5060",
	    "to_name": "",
	    "to_user": "018926798345",
	    "to_host": "219.143.187.139",
	    "to_port": "5060",
	    "contact_name": "1101385",
	    "contact_user": "1101385",
	    "contact_host": "192.168.6.24",
	    "contact_port": "5060",
	    "user_agent": "DonJin SIP Server 3.2.0_i"
	}
	`)
	byeOk := []byte(`
	{
		"event_id": "10020044170",
	  	"event_time": "20201111100903",
	  	"sip": "219.143.187.139",
	    "sport": "5088",
	    "dip": "192.168.6.24",
	    "dport": "5060",
	    "call_id": "04ab01e2d142787@192.168.6.24",
	    "cseq_method": "BYE",
	    "req_method": "",
	    "req_status_code": "200",
	    "req_user": "",
	    "req_host": "",
	    "req_port": "5060",
	    "from_name": "1101385",
	    "from_user": "1101385",
	    "from_host": "192.168.6.24",
	    "from_port": "5060",
	    "to_name": "",
	    "to_user": "018926798345",
	    "to_host": "219.143.187.139",
	    "to_port": "5060",
	    "contact_name": "1101385",
	    "contact_user": "1101385",
	    "contact_host": "192.168.6.24",
	    "contact_port": "5060",
	    "user_agent": ""
	}
	`)

	cdr.ParseInvite200OKMessage([]byte(callid), inviteOk)
	cdrPkt := cdr.ParseBye200OKMsg([]byte(callid), byeOk)
	jsonStr, err := json.Marshal(cdrPkt)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug(jsonStr)
	//consumer.next.Log("", string(jsonStr))
}
