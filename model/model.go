package model

import (
    "bufio"
    . "centnet-cdrrs/adapter/go-cache"
    "centnet-cdrrs/adapter/kafka"
    "centnet-cdrrs/conf"
    "centnet-cdrrs/dao"
    "centnet-cdrrs/library/log"
    "encoding/json"
    "time"
)

var writer *bufio.Writer
var cache *ShardedCache

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

    var cdr dao.VoipRestoredCdr
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

const (
    SipInvite200OK = iota
    SipBye200OK
)

type sipItem struct {
    msgType uint8
}

func DoLine(line interface{}) {

    // 解析sip报文
    pkt := NewSip()
    if nil != pkt.parse(line) {
        pkt.Free()
        return
    }

    if pkt.CseqMethod == "INVITE" && pkt.ReqStatusCode == 200 {
        v, ok := cache.Get(pkt.CallId)
        if !ok {
            cache.Set(pkt.CallId, sipItem{SipInvite200OK}, time.Hour)
            // insert database
            cdrRestore(pkt, pkt)
        } else if v.(sipItem).msgType == SipBye200OK {
            cache.Delete(pkt.CallId)
            // update database
        }
    } else if pkt.CseqMethod == "BYE" && pkt.ReqStatusCode == 200 {
        v, ok := cache.Get(pkt.CallId)
        if !ok {
            cache.Set(pkt.CallId, sipItem{msgType: SipBye200OK}, time.Hour)
            // insert database
            cdrRestore(pkt, pkt)
        } else if v.(sipItem).msgType == SipInvite200OK {
            cache.Delete(pkt.CallId)
            // update database
        }
    } else {
        log.Debug("no handler for else condition")
    }

    //if pkt.CseqMethod == "INVITE" && pkt.ReqStatusCode == 200 {
    //   if pkt.doInvite200OKMessage() == nil {
    //       return
    //   }
    //} else if pkt.CseqMethod == "BYE" && pkt.ReqStatusCode == 200 {
    //   if pkt.doBye200OKMessage() == nil {
    //       return
    //   }
    //} else {
    //   log.Debug("no handler for else condition")
    //}

    pkt.Free()
}

func Init() error {
    /* 还原的话单数据交给诈骗分析模型 */
    producerConfig := conf.Conf.Kafka.FraudModelProducer
    if producerConfig.Enable {
        var err error
        fraudAnalysisModel, err = kafka.NewProducer(producerConfig)
        if err != nil {
            log.Error(err)
            return err
        }
        fraudAnalysisModel.Run()
    }

    cache = NewShardedCache(time.Hour, time.Second*10, 1024)

    return nil
}
