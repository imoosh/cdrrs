package model

import (
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"encoding/json"
	"time"
)

type SipAnalyticPacket struct {
	Id            uint64 `json:"id"`
	EventId       string `json:"eventId"`
	EventTime     string `json:"eventTime"`
	Sip           string `json:"sip"`
	Sport         int    `json:"sport"`
	Dip           string `json:"dip"`
	Dport         int    `json:"dport"`
	CallId        string `json:"callId"`
	CseqMethod    string `json:"cseqMethod"`
	ReqMethod     string `json:"reqMethod"`
	ReqStatusCode int    `json:"reqStatusCode"`
	ReqUser       string `json:"reqUser"`
	ReqHost       string `json:"reqHost"`
	ReqPort       int    `json:"reqPort"`
	FromName      string `json:"fromName"`
	FromUser      string `json:"fromUser"`
	FromHost      string `json:"fromHost"`
	FromPort      int    `json:"fromPort"`
	ToName        string `json:"toName"`
	ToUser        string `json:"toUser"`
	ToHost        string `json:"toHost"`
	ToPort        int    `json:"toPort"`
	ContactName   string `json:"contactName"`
	ContactUser   string `json:"contactUser"`
	ContactHost   string `json:"contactHost"`
	ContactPort   int    `json:"contactPort"`
	UserAgent     string `json:"userAgent"`
}

func HandleInvite200OKMessage(key, value string) {
	//	invite200OK插入redis
	redis.RedisConn.PutWithExpire(key, value, redis.RedisConn.Conf.CacheExpire)
	log.Debugf("Insert redis ok: %s -- %s", key, value)
}

func HandleBye200OKMsg(key string, bye200OKMsg SipAnalyticPacket) *dao.VoipRestoredCdr {
	var err error

	//根据BYE 200OK包的call_id从redis获取对应的INVITE 200OK的包
	invite200OKRawMsg, notGet := redis.RedisConn.Get(key)
	if notGet != nil {
		log.Error("get invite200ok failed")
		return nil
	}

	var invite200OKMsg SipAnalyticPacket
	if err = json.Unmarshal([]byte(invite200OKRawMsg), &invite200OKMsg); err != nil {
		log.Error(err)
		return nil
	}

	//查询成功后删除redis中INVITE 200OK包
	redis.RedisConn.Delete(key)

	//获取被叫归属地
	calleeAttribution := dao.GetPositionByPhoneNum(invite200OKMsg.ToUser)
	connectTime, err := time.Parse(invite200OKMsg.EventTime, "2006-01-02 15:04:05")
	if err != nil {
		log.Errorf("time.Parse error: %s", invite200OKMsg.EventTime)
		return nil
	}
	disconnectTime, err := time.Parse(invite200OKMsg.EventTime, "2006-01-02 15:04:05")
	if err != nil {
		log.Errorf("time.Parse error: %s", invite200OKMsg.EventTime)
		return nil
	}

	//填充话单字段信息
	cdr := dao.VoipRestoredCdr{
		CallId:         invite200OKMsg.CallId,
		CallerIp:       invite200OKMsg.Dip,
		CallerPort:     invite200OKMsg.Dport,
		CalleeIp:       invite200OKMsg.Sip,
		CalleePort:     invite200OKMsg.Sport,
		CallerNum:      invite200OKMsg.FromUser,
		CalleeNum:      invite200OKMsg.ToUser,
		CalleeDevice:   invite200OKMsg.UserAgent,
		CalleeProvince: calleeAttribution.Province,
		CalleeCity:     calleeAttribution.City,
		ConnectTime:    connectTime,
		DisconnectTime: disconnectTime,
	}

	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if invite200OKMsg.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent
	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}

	cdr.Duration = int(cdr.ConnectTime.Sub(cdr.DisconnectTime).Seconds())
	cdr.FraudType = ""

	dao.InsertCDR(&cdr)

	return &cdr
}
