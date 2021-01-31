package service

import (
    "centnet-cdrrs/common/cache/local"
    "centnet-cdrrs/common/kafka"
    "centnet-cdrrs/common/log"
    "centnet-cdrrs/conf"
    "centnet-cdrrs/dao"
    "centnet-cdrrs/model"
    "centnet-cdrrs/service/adapters/file"
    "centnet-cdrrs/service/cdr"
    "centnet-cdrrs/service/voip"
    "encoding/json"
    "fmt"
    "time"
)

var (
    _cdrCh = make(chan *dao.VoipCDR)
)

type Service struct {
    c          *conf.Config
    parser     *file.Parser
    dao        *dao.Dao
    fraudModel *kafka.Producer
    mc         *local.ShardedCache
    cdrPro     *cdr.CDRProducer
    synth      *Synth
    cdrs       []*dao.VoipCDR
}

func New(c *conf.Config) (svr *Service, err error) {

    // 话单合成模块
    cdrPro := cdr.NewCDRProducer(c.CDR)
    cdrPro.SetDispatchFunc(handleCDRCallback)
    cdrPro.Dispatch()

    // kafka交互模块
    var fm *kafka.Producer
    if c.Kafka.FraudModel.Enable {
        fm, err = kafka.NewProducer(c.Kafka.FraudModel)
        if err != nil {
            log.Error(err)
            panic(err)
        }
        fm.Run()
    }

    // 预读号码归属地缓存
    dao := dao.New(c)
    if err = dao.CachePosition(); err != nil {
        log.Error(err)
        panic(err)
    }

    // 本地缓存引擎
    //mc := local.NewShardedCache(time.Duration(c.CDR.CachedLife), time.Second*30, 1024)
    //mc.OnEvicted(func(k string, v interface{}) {
    //    cdrPro.GenExpiredCDR(k, v.(*model.SipItem))
    //})

    // 初始化数据解析器
    svr = &Service{
        c:          c,
        parser:     file.NewParser(c.FileParser),
        cdrPro:     cdrPro,
        dao:        dao,
        fraudModel: fm,
    }

    synth := newSynthesizer(svr)
    synth.OnExpired(func(args interface{}) {
        items := args.(map[string]*model.SipItem)
        for k, v := range items {
            if v != nil {
                cdrPro.GenExpiredCDR(k, v)
            }
        }
    })
    synth.Run()
    svr.synth = synth

    // 启动话单处理服务
    go svr.handleCDR()

    // 最后启动数据解析器
    svr.parser.SetParseFunc(svr.DoLine)
    svr.parser.Run()

    return svr, nil
}

func (srv *Service) DoLine(line interface{}) {

    // 解析sip报文
    var sip model.SipPacket
    if nil != voip.ParseSipPacket(line, &sip) {
        return
    }

    // invite-200ok消息
    if sip.CseqMethod == "INVITE" && sip.ReqStatusCode == 200 {
        invItem := model.NewSipItemFromSipPacket(&sip)
        invItem.Type = model.SipStatusInvite200OK
        invItem.CalleeDevice = sip.UserAgent
        invItem.ConnectTime = sip.EventTime

        srv.synth.Input(invItem)
    } else if sip.CseqMethod == "BYE" && sip.ReqStatusCode == 200 {
        byeItem := model.NewSipItemFromSipPacket(&sip)
        byeItem.Type = model.SipStatusBye200OK
        byeItem.DisconnectTime = sip.EventTime

        srv.synth.Input(byeItem)
    } else {
        log.Debug("no handler for else condition")
    }
}

func handleCDRCallback(vc interface{}) {
    _cdrCh <- vc.(*dao.VoipCDR)
}

func (srv *Service) handleCDR() {
    var (
        duration     time.Duration
        procDBTime   time.Time
        curTableName string
    )

    duration = time.Duration(srv.c.CDR.CdrFlushPeriod)
    ticker := time.NewTicker(duration)

    for {
        select {
        case c := <-_cdrCh:
            srv.pushCDRToKafka(c)

            if c.TableName != curTableName {
                srv.dao.CreateTable(c.TableName)
                // 不是第一次建表，将缓存话单先入旧库
                if len(curTableName) != 0 {
                    srv.flushCDRToDB(curTableName, srv.cdrs)
                    procDBTime = time.Now()
                }
                curTableName = c.TableName
            }

            srv.cdrs = append(srv.cdrs, c)
            if len(srv.cdrs) == srv.c.CDR.MaxFlushCap {
                srv.flushCDRToDB(curTableName, srv.cdrs)
                procDBTime = time.Now()
            }
        case <-ticker.C:
            // 有一段时间没刷新数据库
            now := time.Now()
            if now.Sub(procDBTime) >= duration {
                srv.flushCDRToDB(curTableName, srv.cdrs)
                procDBTime = now
            }
        }
    }
}

func (srv *Service) clearCDRs() {
    for _, x := range srv.cdrs {
        srv.cdrPro.P.Free(x)
    }
    srv.cdrs = srv.cdrs[:0]
}

func (srv *Service) pushCDRToKafka(c *dao.VoipCDR) {
    // 推送至诈骗分析模型
    if srv.c.Kafka.FraudModel.Enable {
        cdrStr, err := json.Marshal(c)
        if err != nil {
            log.Errorf("json.Marshal error: ", err)
            return
        }

        srv.fraudModel.Log(fmt.Sprintf("%d", c.Id), string(cdrStr))
    }
}

func (srv *Service) flushCDRToDB(tblName string, cdrs []*dao.VoipCDR) {
    if len(cdrs) == 0 {
        return
    }
    srv.dao.MultiInsertCDR(tblName, cdrs)
    srv.clearCDRs()
}

//func writeCDRToTxt(m *kafka.ConsumerGroupMember, k, v interface{}) {
//    m.TotalCount++
//    m.TotalBytes = m.TotalBytes + uint64(len(v.([]byte)))
//
//    _, err := writer.WriteString(string(v.([]byte)) + "\n")
//    if err != nil {
//        log.Error(err)
//    }
//}
//
//func writeCDRToDB(m *kafka.ConsumerGroupMember, k, v interface{}) {
//    m.TotalCount++
//    m.TotalBytes = m.TotalBytes + uint64(len(v.([]byte)))
//
//    var cdr dao.VoipCDR
//    err := json.Unmarshal(v.([]byte), &cdr)
//    if err != nil {
//        log.Error(err)
//        return
//    }
//}
//
//func initWriteCDRKafkaConsumer() {
//    // kafka消费者： 话单写数据库
//    kafka.NewConsumerGroupMember(&kafka.ConsumerConfig{
//        Broker:              "192.168.1.205:9092",
//        Topic:               "cdr",
//        Group:               "cdr-db",
//        GroupMembers:        1,
//        FlowRateFlushPeriod: 3}, "cdr-write-db_01", writeCDRToDB)
//}
