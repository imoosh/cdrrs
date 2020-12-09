package model

import (
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"time"
)

var fraudAnalysisModel *kafka.Producer

var redisResultCount = 0

func HandleRedisResult(unit redis.DelayHandleUnit, result redis.CmdResult) {
	redisResultCount++
	if redisResultCount%10000 == 0 {
		log.Debug("redis result count:", redisResultCount)
	}

	if result.Err != nil {
		log.Error(result.Err)
		return
	}
	if unit.Func == nil {
		return
	}

	switch result.Value.(type) {
	case nil:
		sipmsg := unit.Args.(AnalyticSipPacket)
		if sipmsg.CseqMethod == "INVITE" {
			log.Error("### cannot find bye 200 ok:", sipmsg.CallId)
		} else if sipmsg.CseqMethod == "BYE" {
			log.Error("### cannot find invite 200 ok:", sipmsg.CallId)
		} else {
			log.Error("### invalid sip message")
		}
	case string:
		log.Debug("### ", result.Value.(string))
	case []byte:
		var pkt AnalyticSipPacket
		err := json.Unmarshal(result.Value.([]byte), &pkt)
		if err != nil {
			log.Error(err)
		}

		// 合成话单
		if pkt.CseqMethod == "INVITE" && unit.Args.(AnalyticSipPacket).CseqMethod == "BYE" {
			// 实际调用cdrRestore
			//unit.Func(pkt, unit.Args)
			cdrRestore(pkt, unit.Args)
		} else if pkt.CseqMethod == "BYE" && unit.Args.(AnalyticSipPacket).CseqMethod == "INVITE" {
			// 实际调用cdrRestore
			//unit.Func(unit.Args, pkt)
			cdrRestore(unit.Args, pkt)
		} else {
			log.Error("### error message type: ", string(result.Value.([]byte)))
			sipmsg := unit.Args.(AnalyticSipPacket)
			sipstr, _ := json.Marshal(&sipmsg)
			log.Error("### error message pair: ", string(sipstr))

			//if sipmsg.CseqMethod == "INVITE" {
			//	log.Error("### cannot find bye message by call-id: ", sipmsg.CallId)
			//} else if sipmsg.CseqMethod == "BYE" {
			//	log.Error("### cannot find invite message by call-id: ", sipmsg.CallId)
			//}
		}
	default:
		log.Fatal("### ", result.Value)
	}
}

func cdrRestore(i, b interface{}) interface{} {
	invite, bye := i.(AnalyticSipPacket), b.(AnalyticSipPacket)

	//consumer.TotalCount++
	//consumer.TotalBytes = consumer.TotalBytes + uint64(len(value.([]byte)))

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
	cdr := dao.VoipRestoredCdr{
		CallId:         invite.CallId,
		Uuid:           uuid.NewV4().String(),
		CallerIp:       invite.Dip,
		CallerPort:     invite.Dport,
		CalleeIp:       invite.Sip,
		CalleePort:     invite.Sport,
		CallerNum:      invite.FromUser,
		CalleeNum:      invite.ToUser,
		CalleeDevice:   invite.UserAgent,
		CalleeProvince: "",
		CalleeCity:     "",
		ConnectTime:    connectTime.Unix(),
		DisconnectTime: disconnectTime.Unix(),
		Duration:       0,
		FraudType:      "",
		CreateTime:     now.Format("2006-01-02 15:04:05"),
		CreateTimeX:    now,
	}

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
	cdr.CalleeNum = calleeInfo.Number
	cdr.CalleeProvince = calleeInfo.Pos.Province
	cdr.CalleeCity = calleeInfo.Pos.City

	// 推送至诈骗分析模型
	if conf.Conf.Kafka.FraudModelProducer.Enable {
		cdrStr, err := json.Marshal(&cdr)
		if err != nil {
			log.Errorf("json.Marshal error: ", err)
			return nil
		}
		fraudAnalysisModel.Log(invite.CallId, string(cdrStr))
	}

	//插入话单数据库
	dao.LogCDR(&cdr)

	return nil
}

func Init() error {
	/* 还原的话单数据交给诈骗分析模型 */
	producerConfig := conf.Conf.Kafka.FraudModelProducer
	if producerConfig.Enable {
		fraudAnalysisModel, err := kafka.NewProducer(producerConfig)
		if err != nil {
			log.Error(err)
			return err
		}
		fraudAnalysisModel.Run()
	}

	return nil
}
