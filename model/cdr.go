package model

import (
    "centnet-cdrrs/adapter/kafka"
    "centnet-cdrrs/adapter/redis"
    "centnet-cdrrs/conf"
    "centnet-cdrrs/dao"
    "centnet-cdrrs/library/log"
    "encoding/json"
    "go.uber.org/atomic"
    "time"
)

var (
    sn                 atomic.Int32
    fraudAnalysisModel *kafka.Producer
)

func DoRedisResult(unit redis.DelayHandleUnit, result redis.CmdResult) {

    // 统一回收AnalyticSipPacket, lpkt：本地缓存，rpkt：redis缓存
    var lpkt, rpkt *AnalyticSipPacket
    if unit.Args != nil {
        lpkt = unit.Args.(*AnalyticSipPacket)
    }

    switch result.Value.(type) {
    case nil:
        if lpkt == nil {
            break
        }

        if lpkt.CseqMethod == "INVITE" {
            log.Error("cannot find BYE-200OK: ", lpkt.CallId)
        } else if lpkt.CseqMethod == "BYE" {
            log.Error("cannot find INVITE-200OK: ", lpkt.CallId)
        } else {
            log.Error("invalid sip message")
        }

        // 同一sip会话中，INVITE-200OK与BYE-200OK，会严格控制只有一个包缓存至Redis。
        // 在正常情况下，在每条Redis的Key超时删除之前，一定会GET到数据的，（例外: DoLine接口分别同时处理两种包时）
        // 但由于多routine任务调度，GET命令会在SET命令之前发送至Redis服务器，导致GET失败。
        // 所以在这里尝试l从新GET数据，并且保证同一Key只尝试从新GET一次数据（使用GetAgain标记）
        //
        // 使用go func()异步处理原因：
        // 1、DoRedisResult函数中，range todoQueue处理流程里不能直接或间接使用todoQueue <- x
        // 2、进入从新GET数据流程概率极低，极少开启go func()异步处理，所以不会增加多少系统负担
        if lpkt.GetAgain {
            break
        }
        go func(asp *AnalyticSipPacket) {
            time.Sleep(time.Second * 3)
            asp.GetAgain = true
            redis.AsyncLoad(asp.CallId, redis.DelayHandleUnit{
                Func: cdrRestore,
                Args: asp,
            })
        }(NewSip().Init(lpkt))

    case []byte:
        if lpkt == nil {
            break
        }

        rpkt = NewSip()
        if err := json.Unmarshal(result.Value.([]byte), rpkt); err != nil {
            log.Error(err)
            break
        }
        // 补齐call-id
        rpkt.CallId = lpkt.CallId

        // 获取到后，立即删除缓存
        redis.AsyncDelete(rpkt.CallId)

        // 合成话单
        if rpkt.CseqMethod == "INVITE" && lpkt.CseqMethod == "BYE" {
            // 实际调用cdrRestore
            //unit.Func(pkt, unit.Args)
            cdrRestore(rpkt, lpkt)
        } else if rpkt.CseqMethod == "BYE" && lpkt.CseqMethod == "INVITE" {
            // 实际调用cdrRestore
            //unit.Func(unit.Args, pkt)
            cdrRestore(lpkt, rpkt)
        } else {
            sipStr, _ := json.Marshal(lpkt)
            log.Error("error message type: ", string(result.Value.([]byte)))
            log.Error("error message pair: ", string(sipStr))
        }
    case string:
    default:
        log.Fatal("error redis response: ", result.Value)
    }

    // sip报文本地缓存
    if lpkt != nil {
        lpkt.Free()
    }
    // sip报文redis缓存
    if rpkt != nil {
        rpkt.Free()
    }
}

func cdrId(t time.Time) int64 {
    return int64((t.Hour()*60*60+t.Minute()*60+t.Second())*1e9+t.Nanosecond()/1e3*1e3) + int64(sn.Add(1)%1e3)
}

func cdrRestore(i, b interface{}) interface{} {
    invite, bye := i.(*AnalyticSipPacket), b.(*AnalyticSipPacket)

    var err error
    connectTime, err := time.ParseInLocation("20060102150405", invite.EventTime, time.Local)
    if err != nil {
        log.Errorf("time.Parse error: %s", invite.EventTime)
        return nil
    }
    disconnectTime, err := time.ParseInLocation("20060102150405", bye.EventTime, time.Local)
    if err != nil {
        log.Errorf("time.Parse error: %s", bye.EventTime)
        return nil
    }

    //填充话单字段信息
    now := time.Now()
    cdr := dao.NewCDR()
    cdr.Id = cdrId(now)
    cdr.CallerIp = invite.Dip
    cdr.CallerPort = invite.Sport
    cdr.CalleeIp = invite.Sip
    cdr.CalleePort = invite.Dport
    cdr.CallerNum = invite.FromUser
    cdr.CalleeNum = invite.ToUser
    cdr.CalleeDevice = invite.UserAgent
    cdr.CalleeProvince = ""
    cdr.CalleeCity = ""
    cdr.ConnectTime = connectTime.Unix()
    cdr.DisconnectTime = disconnectTime.Unix()
    cdr.Duration = 0
    cdr.FraudType = ""
    cdr.CreateTime = now.Format("2006-01-02 15:04:05")
    cdr.CreateTimeX = now

    //INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
    if invite.Dip == bye.Sip {
        cdr.CallerDevice = bye.UserAgent
    } else {
        //主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
        cdr.CallerDevice = ""
    }
    cdr.Duration = int(disconnectTime.Sub(connectTime).Seconds())

    // 填充原始被叫号码及归属地
    calleeInfo := bye.CalleeInfo
    cdr.CalleeNum = calleeInfo.Num
    cdr.CalleeProvince = calleeInfo.Pos.Province
    cdr.CalleeCity = calleeInfo.Pos.City

    // 推送至诈骗分析模型
    if conf.Conf.Kafka.FraudModelProducer.Enable {
        cdrStr, err := json.Marshal(cdr)
        if err != nil {
            log.Errorf("json.Marshal error: ", err)
            cdr.Free()
            return nil
        }
        fraudAnalysisModel.Log(invite.CallId, string(cdrStr))
    }

    //插入话单数据库
    dao.InsertCDR(cdr)

    return nil
}

func newCDRFromInvite200OK(pkt *AnalyticSipPacket) *dao.VoipRestoredCdr {
    var err error
    connectTime, err := time.ParseInLocation("20060102150405", pkt.EventTime, time.Local)
    if err != nil {
        log.Errorf("time.Parse error: %s", pkt.EventTime)
        return nil
    }

    //填充话单字段信息
    now := time.Now()
    cdr := dao.NewCDR()
    cdr.Id = cdrId(now)
    cdr.CallerIp = pkt.Dip
    cdr.CallerPort = pkt.Sport
    cdr.CalleeIp = pkt.Sip
    cdr.CalleePort = pkt.Dport
    cdr.CallerNum = pkt.FromUser
    cdr.CalleeNum = pkt.ToUser
    cdr.CalleeDevice = pkt.UserAgent
    cdr.CalleeProvince = ""
    cdr.CalleeCity = ""
    cdr.ConnectTime = connectTime.Unix()
    cdr.DisconnectTime = connectTime.Unix()
    cdr.Duration = 0
    cdr.FraudType = ""
    cdr.CreateTime = now.Format("2006-01-02 15:04:05")
    cdr.CreateTimeX = now
    cdr.CalleeDevice = pkt.UserAgent

    // 填充原始被叫号码及归属地
    calleeInfo := pkt.CalleeInfo
    cdr.CalleeNum = calleeInfo.Num
    cdr.CalleeProvince = calleeInfo.Pos.Province
    cdr.CalleeCity = calleeInfo.Pos.City

    // 推送至诈骗分析模型
    if conf.Conf.Kafka.FraudModelProducer.Enable {
        cdrStr, err := json.Marshal(cdr)
        if err != nil {
            log.Errorf("json.Marshal error: ", err)
            cdr.Free()
            return nil
        }
        fraudAnalysisModel.Log(pkt.CallId, string(cdrStr))
    }

    //插入话单数据库
    dao.InsertCDR(cdr)

    return nil
}

func newCDRFromBye200OK(pkt *AnalyticSipPacket) {
    var err error
    connectTime, err := time.ParseInLocation("20060102150405", pkt.EventTime, time.Local)
    if err != nil {
        log.Errorf("time.Parse error: %s", pkt.EventTime)
        return
    }
    disconnectTime, err := time.ParseInLocation("20060102150405", bye.EventTime, time.Local)
    if err != nil {
        log.Errorf("time.Parse error: %s", bye.EventTime)
        return
    }

    //填充话单字段信息
    now := time.Now()
    cdr := dao.NewCDR()
    cdr.Id = cdrId(now)
    cdr.CallerIp = invite.Dip
    cdr.CallerPort = invite.Sport
    cdr.CalleeIp = invite.Sip
    cdr.CalleePort = invite.Dport
    cdr.CallerNum = invite.FromUser
    cdr.CalleeNum = invite.ToUser
    cdr.CalleeDevice = invite.UserAgent
    cdr.CalleeProvince = ""
    cdr.CalleeCity = ""
    cdr.ConnectTime = connectTime.Unix()
    cdr.DisconnectTime = disconnectTime.Unix()
    cdr.Duration = 0
    cdr.FraudType = ""
    cdr.CreateTime = now.Format("2006-01-02 15:04:05")
    cdr.CreateTimeX = now

    //INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
    if invite.Dip == bye.Sip {
        cdr.CallerDevice = bye.UserAgent
    } else {
        //主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
        cdr.CallerDevice = ""
    }
    cdr.Duration = int(disconnectTime.Sub(connectTime).Seconds())

    // 填充原始被叫号码及归属地
    calleeInfo := bye.CalleeInfo
    cdr.CalleeNum = calleeInfo.Num
    cdr.CalleeProvince = calleeInfo.Pos.Province
    cdr.CalleeCity = calleeInfo.Pos.City

    // 推送至诈骗分析模型
    if conf.Conf.Kafka.FraudModelProducer.Enable {
        cdrStr, err := json.Marshal(cdr)
        if err != nil {
            log.Errorf("json.Marshal error: ", err)
            cdr.Free()
            return nil
        }
        fraudAnalysisModel.Log(invite.CallId, string(cdrStr))
    }

    //插入话单数据库
    dao.InsertCDR(cdr)

    return nil
}

func updateCDRConnectTime(id int64, t time.Time) {
    cdr := dao.NewCDR()
    cdr.Id = id
    cdr.ConnectTime = t.Unix()
}

func updateCDRDisconnectTime(id int64, t time.Time) {
    cdr := dao.NewCDR()
    cdr.Id = id
    cdr.DisconnectTime = t.Unix()
}
