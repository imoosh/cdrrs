package model

import (
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var errUnresolvableNumber = errors.New("unresolvable number")

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

type CalleeInfo struct {
	Number string
	Pos    dao.PhonePosition
}

func validatePhoneNumber(num string) bool {
	for _, v := range num {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func ParseCalleeInfo(num string) (CalleeInfo, error) {
	length := len(num)
	if !validatePhoneNumber(num) || length < 11 {
		return CalleeInfo{Number: num, Pos: dao.PhonePosition{}}, errUnresolvableNumber
	}

	var (
		callee = num
		err    error
		pos    dao.PhonePosition
	)

	if length >= 11 {
		callee = callee[length-11:]
		if strings.HasPrefix(callee, "1") {
			// 手机号码归属查询
			if pos, err = dao.GetPositionByMobilePhoneNumber(callee); err == nil {
				return CalleeInfo{Number: callee, Pos: pos}, nil
			}
		} else if strings.HasPrefix(callee, "0") {
			// 座机号码归属查询
			if pos, err = dao.GetPositionByFixedPhoneNumber(callee, -1); err == nil {
				return CalleeInfo{Number: callee, Pos: pos}, nil
			}
		}
	}

	if length >= 12 {
		callee = callee[length-12:]
		if strings.HasPrefix(callee, "0") {
			// 座机号码归属查询
			if pos, err = dao.GetPositionByFixedPhoneNumber(callee, 4); err == nil {
				return CalleeInfo{Number: callee, Pos: pos}, nil
			}
		} else if strings.HasPrefix(callee, "86") {
			// 86xx xxxx xxxx -> 80xx xxxx xxxx -> 0xx xxxx xxxx
			callee = strings.Replace(callee, "86", "80", 1)[1:]
			if pos, err = dao.GetPositionByFixedPhoneNumber(callee, -1); err == nil {
				return CalleeInfo{Number: callee, Pos: pos}, nil
			}
		}
	}

	if length >= 13 {
		callee = callee[length-13:]
		if strings.HasPrefix(callee, "86") {
			// 86xxx xxxx xxxx -> 80xxx xxxx xxxx -> 0xxx xxxx xxxx
			callee = strings.Replace(callee, "86", "80", 1)[1:]
			if pos, err = dao.GetPositionByFixedPhoneNumber(callee, 4); err == nil {
				return CalleeInfo{Number: callee, Pos: pos}, nil
			}
		}
	}

	callee = num[length-11:]
	if strings.HasPrefix(callee, "852") {
		return CalleeInfo{Number: callee, Pos: dao.PhonePosition{Province: "香港", City: "香港"}}, nil
	} else if strings.HasPrefix(callee, "853") {
		return CalleeInfo{Number: callee, Pos: dao.PhonePosition{Province: "澳门", City: "澳门"}}, nil
	}

	return CalleeInfo{Number: num, Pos: pos}, errUnresolvableNumber
}

func HandleInvite200OKMessage(key, value string) {

	//	invite200OK插入redis
	_ = redis.PutWithExpire(key, value, redis.Conf.CacheExpire)
}

func HandleBye200OKMsg(key string, bye200OKMsg SipAnalyticPacket) *dao.VoipRestoredCdr {
	var err error

	//根据BYE 200OK包的call_id从redis获取对应的INVITE 200OK的包
	invite200OKRawMsg, err := redis.Get(key)
	if err == nil && invite200OKRawMsg == "" {
		// 未获取到数据
		return nil
	} else if err != nil {
		log.Error("redis.Get failed")
		return nil
	}

	var invite200OKMsg SipAnalyticPacket
	if err = json.Unmarshal([]byte(invite200OKRawMsg), &invite200OKMsg); err != nil {
		log.Error(err)
		return nil
	}

	//查询成功后删除redis中INVITE 200OK包
	redis.Delete(key)

	connectTime, err := time.ParseInLocation("20060102150405", invite200OKMsg.EventTime, time.Local)
	//connectTime, err := time.Parse(invite200OKMsg.EventTime, "2006-01-02 15:04:05")
	if err != nil {
		log.Errorf("time.Parse error: %s", invite200OKMsg.EventTime)
		return nil
	}
	disconnectTime, err := time.ParseInLocation("20060102150405", bye200OKMsg.EventTime, time.Local)
	//disconnectTime, err := time.Parse(invite200OKMsg.EventTime, "2006-01-02 15:04:05")
	if err != nil {
		log.Errorf("time.Parse error: %s", invite200OKMsg.EventTime)
		return nil
	}

	//log.Debug(connectTime, disconnectTime)

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
		CalleeProvince: "",
		CalleeCity:     "",
		ConnectTime:    connectTime.Format("2006-01-02 15:04:05"),
		DisconnectTime: disconnectTime.Format("2006-01-02 15:04:05"),
	}

	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if invite200OKMsg.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent
	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}

	cdr.Duration = int(disconnectTime.Sub(connectTime).Seconds())
	cdr.FraudType = ""

	return &cdr
}
