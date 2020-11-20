package kafka

import (
	"centnet-cdrrs/adapter/kafka/file"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/cdr"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"encoding/json"
	"fmt"
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
func AnalyzePacket(consumer *Consumer, data interface{}) {

	rtd := file.Parse(string(data.([]byte)))
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

func RestoreCDR(consumer *Consumer, data interface{}) {
	cdr.ParseInvite(data.([]byte))
	cdrpkt := cdr.ParseBye(data.([]byte))
	jsonStr, err := json.Marshal(cdrpkt)
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Println(jsonStr)
	consumer.next.Log("", string(jsonStr))
}
