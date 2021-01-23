package cdr

import (
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/service/voip"
	"go.uber.org/atomic"
	"time"
)

var (
	counter             atomic.Int64
	cdrTablePrefix      = "cdr_"
	defaultDispatchFunc = func(interface{}) {}
)

type CDRProducer struct {
	c            *conf.CDRConfig
	Q            chan *dao.VoipCDR
	P            *pool
	dispatchFunc func(interface{})
}

func NewCDRProducer(c *conf.CDRConfig) *CDRProducer {
	return &CDRProducer{
		c:            c,
		P:            newPool(),
		Q:            make(chan *dao.VoipCDR, c.MaxCacheCap),
		dispatchFunc: defaultDispatchFunc,
	}
}

func (cp *CDRProducer) SetDispatchFunc(fun func(interface{})) {
	cp.dispatchFunc = fun
}

func (cp *CDRProducer) Put(cdr *dao.VoipCDR) {
	cp.Q <- cdr
}

func (cp *CDRProducer) Gen(callId string, item *voip.SipItem) *dao.VoipCDR {
	var (
		err            error
		connectTime    time.Time
		disConnectTime time.Time
		connTimeLen    = len(item.ConnectTime)
		disConnTimeLen = len(item.DisconnectTime)
	)

	if connTimeLen != 0 {
		connectTime, err = time.ParseInLocation("20060102150405", item.ConnectTime, time.Local)
		if err != nil {
			log.Errorf("time.Parse error: %s", item.ConnectTime)
			return nil
		}
	}
	if disConnTimeLen != 0 {
		disConnectTime, err = time.ParseInLocation("20060102150405", item.DisconnectTime, time.Local)
		if err != nil {
			log.Errorf("time.Parse error: %s", item.DisconnectTime)
			return nil
		}
	}
	pos, err := voip.ParsePositionFrom(item.Callee)
	if err != nil {
		log.Errorf("parse position error(%s)", item.Callee)
		return nil
	}

	//填充话单字段信息
	cdr := cp.P.New()
	cdr.CallerIp = item.SrcIP
	cdr.CallerPort = int(item.SrcPort)
	cdr.CalleeIp = item.DestIP
	cdr.CalleePort = int(item.DestPort)
	cdr.CallerNum = item.Caller
	cdr.CalleeNum = item.Callee
	cdr.CallerDevice = ""
	cdr.CalleeDevice = item.CalleeDevice
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
	return cdr
}

func (cp *CDRProducer) GenExpiredCDR(callId string, item *voip.SipItem) {
	if len(item.Caller) != 0 && len(item.Callee) != 0 {
		if cdr := cp.Gen(callId, item); cdr != nil {
			cp.Put(cdr)
		}
	}

	item.Free()
}

func (cp *CDRProducer) Dispatch() {

	go func() {
		for {
			select {
			case cdr := <-cp.Q:
				// 在一个go routine里处理所有话单，确保时间id是顺序产生的
				now := time.Now()
				cdr.Id = cdrId(now)
				cdr.CreateTime = now.Format("2006-01-02 15:04:05")
				cdr.TableName = cdrTable(now, time.Duration(cp.c.CdrTablePeriod))

				cp.dispatchFunc(cdr)
			}
		}
	}()
}

func cdrTable(t time.Time, period time.Duration) string {
	return cdrTablePrefix + t.Truncate(period).Format("20060102150405")
}

func cdrId(t time.Time) int64 {
	return int64((t.Hour()*60*60+t.Minute()*60+t.Second())*1e9+t.Nanosecond()/1e3*1e3) + int64(counter.Add(1)%1e3)
}
