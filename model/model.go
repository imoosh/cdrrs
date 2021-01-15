package model

import (
	"bufio"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	. "centnet-cdrrs/library/cache"
	"centnet-cdrrs/library/kafka"
	"centnet-cdrrs/library/log"
	"encoding/json"
	"sync"
	"time"
)

var (
	writer       *bufio.Writer
	cache        *ShardedCache
	emptySipItem = sipItem{}
	sipItemPool  = sync.Pool{New: func() interface{} { return &sipItem{} }}
)

func writeCDRToTxt(m *kafka.ConsumerGroupMember, k, v interface{}) {
	m.TotalCount++
	m.TotalBytes = m.TotalBytes + uint64(len(v.([]byte)))

	_, err := writer.WriteString(string(v.([]byte)) + "\n")
	if err != nil {
		log.Error(err)
	}
}

func writeCDRToDB(m *kafka.ConsumerGroupMember, k, v interface{}) {
	m.TotalCount++
	m.TotalBytes = m.TotalBytes + uint64(len(v.([]byte)))

	var cdr dao.VoipCDR
	err := json.Unmarshal(v.([]byte), &cdr)
	if err != nil {
		log.Error(err)
		return
	}
	dao.InsertCDR(&cdr)
}

func initWriteCDRKafkaConsumer() {
	// kafka消费者： 话单写数据库
	kafka.NewConsumerGroupMember(&kafka.ConsumerConfig{
		Broker:              "192.168.1.205:9092",
		Topic:               "cdr",
		Group:               "cdr-db",
		GroupMembers:        1,
		FlowRateFlushPeriod: 3}, "cdr-write-db_01", writeCDRToDB)
}

type sipItem struct {
	caller         string
	callee         string
	srcIP          string
	destIP         string
	srcPort        uint16
	destPort       uint16
	connectTime    string
	disconnectTime string
}

func NewSipItem() *sipItem {
	return sipItemPool.Get().(*sipItem)
}

func (pkt *sipItem) Free() {
	*pkt = emptySipItem
	sipItemPool.Put(pkt)
}

func DoLine(line interface{}) {

	// 解析sip报文
	pkt := NewSip()
	if nil != pkt.parse(line) {
		pkt.Free()
		return
	}

	// invite-200ok消息
	if pkt.CseqMethod == "INVITE" && pkt.ReqStatusCode == 200 {
		item := NewSipItem()
		item.caller = pkt.FromUser
		item.callee = pkt.ToUser
		item.srcIP = pkt.Sip
		item.destIP = pkt.Dip
		item.srcPort = uint16(pkt.Sport)
		item.destPort = uint16(pkt.Dport)
		item.connectTime = pkt.EventTime

		if ok, v := cache.AddOrDel(pkt.CallId, item, 0); !ok {
			// 结束时间不为空（bye消息处理过），合并成话单
			if len(v.(*sipItem).disconnectTime) != 0 {
				item.disconnectTime = v.(*sipItem).disconnectTime
				newCDR(pkt.CallId, item)
			}
			item.Free()
		}

	} else if pkt.CseqMethod == "BYE" && pkt.ReqStatusCode == 200 {
		item := NewSipItem()
		item.disconnectTime = pkt.EventTime

		if ok, v := cache.AddOrDel(pkt.CallId, item, 0); !ok {
			// 开始时间不为空（invite消息处理过），合并成话单
			if len(v.(*sipItem).connectTime) != 0 {
				v.(*sipItem).disconnectTime = pkt.EventTime
				newCDR(pkt.CallId, v.(*sipItem))
			}
			item.Free()
		}
	} else {
		log.Debug("no handler for else condition")
	}

	pkt.Free()
}

func Init() error {
	/* 还原的话单数据交给诈骗分析模型 */
	producerConfig := conf.Conf.Kafka.FraudModelProducer
	if producerConfig.Enable {
		var err error
		fraudModel, err = kafka.NewProducer(producerConfig)
		if err != nil {
			log.Error(err)
			return err
		}
		fraudModel.Run()
	}

	cache = NewShardedCache(time.Minute*10, time.Second*10, 1024)
	cache.OnEvicted(semiFinishedCDRHandler)

	return nil
}
