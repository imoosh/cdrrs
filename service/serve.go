package service

import (
    "centnet-cdrrs/common/cache/local"
    "centnet-cdrrs/common/cache/redis"
    "centnet-cdrrs/common/container/pool"
    "centnet-cdrrs/common/kafka"
    "centnet-cdrrs/common/log"
    xtime "centnet-cdrrs/common/time"
    "centnet-cdrrs/conf"
    "centnet-cdrrs/dao"
    "centnet-cdrrs/service/adapters/file"
    "centnet-cdrrs/service/cdr"
    "centnet-cdrrs/service/voip"
    "context"
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
    cdrs       []*dao.VoipCDR
}

func New(c *conf.Config) (s *Service, err error) {

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
    db := dao.New(c)
    if err = db.CachePosition(); err != nil {
        log.Error(err)
        panic(err)
    }

    // 本地缓存引擎
    mc := local.NewShardedCache(time.Duration(c.CDR.CachedLife)*time.Second, time.Second*30, 1024)
    mc.OnEvicted(func(k string, v interface{}) {
        cdrPro.GenExpiredCDR(k, v.(*voip.SipItem))
    })

    // 初始化数据解析器
    s = &Service{
        c:          c,
        parser:     file.NewParser(c.FileParser),
        cdrPro:     cdrPro,
        dao:        db,
        fraudModel: fm,
        mc:         mc,
    }

    // 启动话单处理服务
    go s.handleCDR()

    // 最后启动数据解析器
    s.parser.SetParseFunc(s.DoLine)
    s.parser.Run()

    return s, nil
}

func getConfig() (c *redis.Config) {
    c = &redis.Config{
        Name:         "test",
        Proto:        "tcp",
        Addr:         "192.168.1.205:6379",
        DialTimeout:  xtime.Duration(time.Second),
        ReadTimeout:  xtime.Duration(time.Second * 10),
        WriteTimeout: xtime.Duration(time.Second),
    }
    c.Config = &pool.Config{
        Active:      20,
        Idle:        2,
        IdleTimeout: xtime.Duration(90 * time.Second),
    }
    return
}

func RedisTest() {
    pool := redis.NewPool(getConfig())
    _, _ = pool.Get(context.TODO()).Do("SET", "abc", "123")
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

func (s *Service) DoLine(line interface{}) {

    // 解析sip报文
    var pkt voip.SipPacket
    if nil != voip.ParseSipPacket(line, &pkt) {
        return
    }

    // invite-200ok消息
    if pkt.CseqMethod == "INVITE" && pkt.ReqStatusCode == 200 {
        item := voip.NewSipItem()
        item.Caller = pkt.FromUser
        item.Callee = pkt.ToUser
        item.SrcIP = pkt.Sip
        item.DestIP = pkt.Dip
        item.SrcPort = uint16(pkt.Sport)
        item.DestPort = uint16(pkt.Dport)
        item.ConnectTime = pkt.EventTime

        if ok, v := s.mc.AddOrDel(pkt.CallId, item, 0); !ok {
            // 结束时间不为空（bye消息处理过），合并成话单
            if len(v.(*voip.SipItem).DisconnectTime) != 0 {
                item.DisconnectTime = v.(*voip.SipItem).DisconnectTime
                if c := s.cdrPro.Gen(pkt.CallId, item); c != nil {
                    s.cdrPro.Put(c)
                }
            }
            item.Free()
        }

    } else if pkt.CseqMethod == "BYE" && pkt.ReqStatusCode == 200 {
        item := voip.NewSipItem()
        item.DisconnectTime = pkt.EventTime

        if ok, v := s.mc.AddOrDel(pkt.CallId, item, 0); !ok {
            // 开始时间不为空（invite消息处理过），合并成话单
            if len(v.(*voip.SipItem).ConnectTime) != 0 {
                v.(*voip.SipItem).DisconnectTime = pkt.EventTime
                if c := s.cdrPro.Gen(pkt.CallId, v.(*voip.SipItem)); c != nil {
                    s.cdrPro.Put(c)
                }
            }
            item.Free()
        }
    } else {
        log.Debug("no handler for else condition")
    }
}

func handleCDRCallback(vc interface{}) {
    _cdrCh <- vc.(*dao.VoipCDR)
}

func (s *Service) handleCDR() {
    var d time.Duration
    var curTableName string

    if d = time.Second * time.Duration(s.c.CDR.CdrFlushPeriod); d == 0 {
        d = time.Second * 10
    }
    ticker := time.NewTicker(d)

    for {
        select {
        case c := <-_cdrCh:
            s.pushCDRToKafka(c)

            if c.TableName != curTableName {
                s.dao.CreateTable(c.TableName)
                // 不是第一次建表，将部分话单先入旧库
                if len(curTableName) != 0 {
                    s.flushCDRToDB(curTableName, s.cdrs)
                }
                curTableName = c.TableName
            }

            s.cdrs = append(s.cdrs, c)
            if len(s.cdrs) == s.c.CDR.MaxFlushCap {
                s.flushCDRToDB(curTableName, s.cdrs)
            }
        case <-ticker.C:
            s.flushCDRToDB(curTableName, s.cdrs)
        }
    }
}

func (s *Service) clearCDRs() {
    for _, x := range s.cdrs {
        s.cdrPro.P.Free(x)
    }
    s.cdrs = s.cdrs[:0]
}

func (s *Service) pushCDRToKafka(c *dao.VoipCDR) {
    // 推送至诈骗分析模型
    if s.c.Kafka.FraudModel.Enable {
        cdrStr, err := json.Marshal(c)
        if err != nil {
            log.Errorf("json.Marshal error: ", err)
            return
        }

        s.fraudModel.Log(fmt.Sprintf("%d", c.Id), string(cdrStr))
    }
}

func (s *Service) flushCDRToDB(tblName string, cdrs []*dao.VoipCDR) {
    s.dao.MultiInsertCDR(tblName, cdrs)
    s.clearCDRs()
}
