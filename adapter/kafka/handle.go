package kafka

import (
	"centnet-cdrrs/adapter/kafka/file"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model"
	"centnet-cdrrs/prot/sip"
	"encoding/json"
	"strconv"
	"time"
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

func atoi(s string, n int) (int, error) {
	if len(s) == 0 {
		return n, nil
	}

	return strconv.Atoi(s)
}

var count int

// 解析sip报文
func AnalyzePacket(consumer *Consumer, key, value interface{}) {

	count++
	if count%100000 == 0 {
		log.Debug(time.Now())
	}

	//t := time.Now()
	rtd := file.Parse(string(value.([]byte)))
	if rtd == nil {
		return
	}
	sipMsg := sip.Parse([]byte(rtd.ParamContent))

	//sip.PrintSipStruct(&sipMsg)

	pkt := model.SipAnalyticPacket{
		EventId:       rtd.EventId,
		EventTime:     rtd.EventTime,
		Sip:           rtd.SaddrV4,
		Sport:         0,
		Dip:           rtd.DaddrV4,
		Dport:         0,
		CallId:        string(sipMsg.CallId.Value),
		CseqMethod:    string(sipMsg.Cseq.Method),
		ReqMethod:     string(sipMsg.Req.Method),
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
	if pkt.Sport, err = atoi(rtd.Sport, 0); err != nil {
		log.Errorf("cannot convert %s to an integer: %v", rtd.Sport, rtd)
		return
	}
	if pkt.Dport, err = atoi(rtd.Dport, 0); err != nil {
		log.Errorf("cannot convert %s to an integer", rtd.Dport)
		return
	}
	if pkt.ReqStatusCode, err = atoi(string(sipMsg.Req.StatusCode), 0); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.Req.StatusCode)
		return
	}
	if pkt.ReqPort, err = atoi(string(sipMsg.Req.Port), 5060); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.Req.Port)
		return
	}
	if pkt.FromPort, err = atoi(string(sipMsg.From.Port), 5060); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.From.Port)
		return
	}
	if pkt.ToPort, err = atoi(string(sipMsg.To.Port), 5060); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.To.Port)
		return
	}
	if pkt.ContactPort, err = atoi(string(sipMsg.Contact.Port), 5060); err != nil {
		log.Errorf("cannot convert %s to an integer", sipMsg.Contact.Port)
		return
	}

	jsonStr, err := json.Marshal(pkt)
	if err != nil {
		log.Error(err)
		return
	}

	consumer.next.Log(string(sipMsg.CallId.Value), string(jsonStr))
	//log.Debug("AnalyzePacket duration: ", time.Since(t))
}

var count1 int

func RestoreCDR(consumer *Consumer, key, value interface{}) {

	count1++
	if count1%100000 == 0 {
		log.Debug(time.Now())
	}

	var sipMsg model.SipAnalyticPacket
	err := json.Unmarshal(value.([]byte), &sipMsg)
	if err != nil {
		log.Error(err)
		return
	}

	k, v := string(key.([]byte)), string(value.([]byte))
	//log.Debug("KEY: ", k, ", SIPMSG: ...")

	if sipMsg.CseqMethod == "INVITE" && sipMsg.ReqStatusCode == 200 {
		model.HandleInvite200OKMessage(k, v)
	} else if sipMsg.CseqMethod == "BYE" && sipMsg.ReqStatusCode == 200 {
		cdrPkt := model.HandleBye200OKMsg(k, sipMsg)

		cdrStr, err := json.Marshal(&cdrPkt)
		if err != nil {
			log.Errorf("json.Marshal error: ", err)
			return
		}

		log.Debug(string(cdrStr))

		// 推送诈骗分析模型
		consumer.next.Log(k, string(cdrStr))

		// 插入话单数据库
		//dao.InsertCDR(cdrPkt)
		dao.LogCDR(cdrPkt)
	} else {
		log.Debug("no handler for else condition")
	}
	//log.Debug("RestoreCDR duration: ", time.Since(t))
}
