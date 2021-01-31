package model

import (
	"centnet-cdrrs/common/log"
	"go.uber.org/atomic"
	"sync"
	"time"
)

type VoipCDR struct {
	Id             int64  `json:"id"`
	CallerIp       string `json:"callerIp"`
	CallerPort     int    `json:"callerPort"`
	CalleeIp       string `json:"calleeIp"`
	CalleePort     int    `json:"calleePort"`
	CallerNum      string `json:"callerNum"`
	CalleeNum      string `json:"calleeNum"`
	CallerDevice   string `json:"callerDevice"`
	CalleeDevice   string `json:"calleeDevice"`
	CalleeProvince string `json:"calleeProvince"`
	CalleeCity     string `json:"calleeCity"`
	ConnectTime    int64  `json:"connectTime"`
	DisconnectTime int64  `json:"disconnectTime"`
	Duration       int    `json:"duration"`
	FraudType      string `json:"fraudType"`
	CreateTime     string `json:"createTime"`
	TableName      string `json:"tableName" orm:"-"`
}

var (
	emptyCDR       = VoipCDR{}
	cdrTablePrefix = "cdr_"
	counter        atomic.Int64
	cdrPool        = &sync.Pool{New: func() interface{} { return &VoipCDR{} }}
)

func NewVoipCDR() *VoipCDR {
	cdr := cdrPool.Get().(*VoipCDR)
	*cdr = emptyCDR
	return cdr
}

func (cdr *VoipCDR) Free() {
	cdrPool.Put(cdr)
}

func NewCDR(callId string, item *SipCache) *VoipCDR {
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
	var pos PhonePosition
	if err := pos.Parse(item.Callee); err != nil {
		log.Errorf("parse position error(%s)", item.Callee)
		return nil
	}

	//填充话单字段信息
	cdr := NewVoipCDR()
	cdr.CalleeProvince = pos.Province
	cdr.CalleeCity = pos.City
	cdr.CallerIp = item.SrcIP
	cdr.CallerPort = int(item.SrcPort)
	cdr.CalleeIp = item.DestIP
	cdr.CalleePort = int(item.DestPort)
	cdr.CallerNum = item.Caller
	cdr.CalleeNum = item.Callee
	cdr.CallerDevice = ""
	cdr.CalleeDevice = item.CalleeDevice

	if connTimeLen != 0 {
		cdr.ConnectTime = connectTime.Unix()
	}
	if disConnTimeLen != 0 {
		cdr.DisconnectTime = disConnectTime.Unix()
	}
	if connTimeLen != 0 && disConnTimeLen != 0 {
		cdr.Duration = int(disConnectTime.Unix() - connectTime.Unix())
	}

	//插入话单数据库
	return cdr
}

func NewExpiredCDR(callId string, item *SipCache) *VoipCDR {
	if len(item.Caller) != 0 && len(item.Callee) != 0 {
		return NewCDR(callId, item)
	}
	return nil
}

func (vc *VoipCDR) SetCDRTable(t time.Time, period time.Duration) {
	vc.TableName = cdrTablePrefix + t.Truncate(period).Format("20060102150405")
}

func (vc *VoipCDR) SetCDRId(t time.Time) {
	vc.Id = int64((t.Hour()*60*60+t.Minute()*60+t.Second())*1e9+t.Nanosecond()/1e3*1e3) + int64(counter.Add(1)%1e3)
}

func (vc *VoipCDR) SetCreateTime(t time.Time) {
	vc.CreateTime = t.Format("2006-01-02 15:04:05")
}
