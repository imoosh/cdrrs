package model

import (
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/kafka"
	"centnet-cdrrs/library/log"
	"encoding/json"
	"fmt"
	"go.uber.org/atomic"
	"time"
)

var (
	cdrs       = make(chan *dao.VoipCDR)
	counter    atomic.Int32
	fraudModel *kafka.Producer
)

func init() {
	go func() {
		for {
			select {
			case cdr := <-cdrs:
				// 在一个go routine里处理所有话单，确保时间id是顺序产生的
				now := time.Now()
				cdr.Id = cdrId(now)
				cdr.CreateTime = now.Format("2006-01-02 15:04:05")
				cdr.CreateTimeX = now

				// 插入数据库
				dao.InsertCDR(cdr)

				// 推送至诈骗分析模型
				if conf.Conf.Kafka.FraudModelProducer.Enable {
					cdrStr, err := json.Marshal(cdr)
					if err != nil {
						log.Errorf("json.Marshal error: ", err)
						cdr.Free()
						return
					}
					fraudModel.Log(fmt.Sprintf("%d", cdr.Id), string(cdrStr))
				}
			}
		}
	}()
}

func cdrId(t time.Time) int64 {
	return int64((t.Hour()*60*60+t.Minute()*60+t.Second())*1e9+t.Nanosecond()/1e3*1e3) + int64(counter.Add(1)%1e3)
}

func newCDR(callId string, pkt *sipItem) {
	var (
		err            error
		connectTime    time.Time
		disConnectTime time.Time
		connTimeLen    = len(pkt.connectTime)
		disConnTimeLen = len(pkt.disconnectTime)
	)

	if connTimeLen != 0 {
		connectTime, err = time.ParseInLocation("20060102150405", pkt.connectTime, time.Local)
		if err != nil {
			log.Errorf("time.Parse error: %s", pkt.connectTime)
			return
		}
	}
	if disConnTimeLen != 0 {
		disConnectTime, err = time.ParseInLocation("20060102150405", pkt.disconnectTime, time.Local)
		if err != nil {
			log.Errorf("time.Parse error: %s", pkt.disconnectTime)
			return
		}
	}
	pos, err := parsePositionFrom(pkt.callee)
	if err != nil {
		log.Errorf("parse position error(%s)", pkt.callee)
		return
	}

	//填充话单字段信息
	cdr := dao.NewCDR()
	cdr.CallerIp = pkt.srcIP
	cdr.CallerPort = int(pkt.srcPort)
	cdr.CalleeIp = pkt.destIP
	cdr.CalleePort = int(pkt.destPort)
	cdr.CallerNum = pkt.caller
	cdr.CalleeNum = pkt.callee
	cdr.CallerDevice = ""
	cdr.CalleeDevice = ""
	cdr.FraudType = ""

	if connTimeLen != 0 {
		cdr.ConnectTime = connectTime.Unix()
	}
	if disConnTimeLen != 0 {
		cdr.DisconnectTime = disConnectTime.Unix()
	}
	if connTimeLen != 0 && disConnTimeLen != 0 {
		cdr.Duration = int(disConnectTime.Unix() - connectTime.Unix())
	}

	// 填充原始被叫号码及归属地
	cdr.CalleeProvince = pos.Province
	cdr.CalleeCity = pos.City

	//插入话单数据库
	cdrs <- cdr
}

func newExpiredCDR(callId string, item *sipItem) {
	if len(item.caller) != 0 && len(item.callee) != 0 {
		newCDR(callId, item)
	}

	item.Free()
}

func semiFinishedCDRHandler(k string, v interface{}) {
	newExpiredCDR(k, v.(*sipItem))
}
