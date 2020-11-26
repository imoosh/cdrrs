package model

import (
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/adapter/kafka/file"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"strconv"
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

	CalleeInfo CalleeInfo
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
		Uuid:           uuid.NewV4().String(),
		CallerIp:       invite200OKMsg.Dip,
		CallerPort:     invite200OKMsg.Dport,
		CalleeIp:       invite200OKMsg.Sip,
		CalleePort:     invite200OKMsg.Sport,
		CallerNum:      invite200OKMsg.FromUser,
		CalleeNum:      invite200OKMsg.ToUser,
		CalleeDevice:   invite200OKMsg.UserAgent,
		CalleeProvince: "",
		CalleeCity:     "",
		ConnectTime:    connectTime.Unix(),
		DisconnectTime: disconnectTime.Unix(),
		Duration:       0,
		FraudType:      "",
		CreateTime:     time.Now().Format("2006-01-02 15:04:05"),
	}

	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if invite200OKMsg.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent
	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}
	cdr.Duration = int(disconnectTime.Sub(connectTime).Seconds())

	return &cdr
}

func atoi(s string, n int) (int, error) {
	if len(s) == 0 {
		return n, nil
	}

	return strconv.Atoi(s)
}

// 解析sip报文
func AnalyzePacket(consumer *kafka.Consumer, key, value interface{}) {
	const FlowRateDuration = 100000

	// 计算流速，每10w计算一次
	consumer.Count++
	if consumer.Count%FlowRateDuration == 0 {
		t := time.Now()
		log.Debug("voip packets analysis rate: %.fpps", FlowRateDuration/t.Sub(consumer.Time).Seconds())
		consumer.Time = t
	}

	rtd := file.Parse(string(value.([]byte)))
	if rtd == nil {
		return
	}
	sipMsg := sip.Parse([]byte(rtd.ParamContent))

	//sip.PrintSipStruct(&sipMsg)

	pkt := SipAnalyticPacket{
		EventId:       rtd.EventId,
		EventTime:     rtd.EventTime,
		Sip:           rtd.SaddrV4,
		Sport:         0,
		Dip:           rtd.DaddrV4,
		Dport:         0,
		CallId:        string(sipMsg.CallId.Value),
		CseqMethod:    string(sipMsg.Cseq.Method),
		ReqMethod:     string(sipMsg.Req.Method),
		ReqStatusCode: 0,
		ReqUser:       string(sipMsg.Req.User),
		ReqHost:       string(sipMsg.Req.Host),
		ReqPort:       0,
		FromName:      string(sipMsg.From.Name),
		FromUser:      string(sipMsg.From.User),
		FromHost:      string(sipMsg.From.Host),
		FromPort:      0,
		ToName:        string(sipMsg.To.Name),
		ToUser:        string(sipMsg.To.User),
		ToHost:        string(sipMsg.To.Host),
		ToPort:        0,
		ContactName:   string(sipMsg.Contact.Name),
		ContactUser:   string(sipMsg.Contact.User),
		ContactHost:   string(sipMsg.Contact.Host),
		ContactPort:   0,
		UserAgent:     string(sipMsg.Ua.Value),
	}

	// 没有call-id、cseq.method、直接丢弃
	if len(pkt.CallId) == 0 || len(pkt.CseqMethod) == 0 || len(pkt.ToUser) == 0 {
		return
	}

	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
	var err error
	pkt.CalleeInfo, err = ParseCalleeInfo(pkt.ToUser)
	if err != nil {
		log.Debug("DIRTY-DATA:", string(value.([]byte)))
		return
	}

	if pkt.Sport, err = atoi(rtd.Sport, 0); err != nil {
		return
	}
	if pkt.Dport, err = atoi(rtd.Dport, 0); err != nil {
		return
	}
	if pkt.ReqStatusCode, err = atoi(string(sipMsg.Req.StatusCode), 0); err != nil {
		return
	}
	if pkt.ReqPort, err = atoi(string(sipMsg.Req.Port), 5060); err != nil {
		return
	}
	if pkt.FromPort, err = atoi(string(sipMsg.From.Port), 5060); err != nil {
		return
	}
	if pkt.ToPort, err = atoi(string(sipMsg.To.Port), 5060); err != nil {
		return
	}
	if pkt.ContactPort, err = atoi(string(sipMsg.Contact.Port), 5060); err != nil {
		return
	}

	jsonStr, err := json.Marshal(pkt)
	if err != nil {
		log.Error(err)
		return
	}

	consumer.Next.Log(string(sipMsg.CallId.Value), string(jsonStr))
	log.Debug(string(jsonStr))
}

var count1 int

func RestoreCDR(consumer *kafka.Consumer, key, value interface{}) {

	count1++
	if count1%100000 == 0 {
		log.Debug(time.Now())
	}

	// 反序列化sip报文字段数据
	var sipMsg SipAnalyticPacket
	err := json.Unmarshal(value.([]byte), &sipMsg)
	if err != nil {
		log.Error(err)
		return
	}

	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
	//calleeInfo, err := ParseCalleeInfo(sipMsg.ToUser)
	//if err != nil {
	//    log.Debug("DIRTY-DATA:", string(value.([]byte)))
	//    return
	//}

	// call-id与sip报文原始数据
	k, v := string(key.([]byte)), string(value.([]byte))
	if sipMsg.CseqMethod == "INVITE" && sipMsg.ReqStatusCode == 200 {
		HandleInvite200OKMessage(k, v)
	} else if sipMsg.CseqMethod == "BYE" && sipMsg.ReqStatusCode == 200 {
		cdrPkt := HandleBye200OKMsg(k, sipMsg)
		if cdrPkt == nil {
			return
		}

		// 填充原始被叫号码及归属地
		calleeInfo := sipMsg.CalleeInfo
		cdrPkt.CalleeNum = calleeInfo.Number
		cdrPkt.CalleeProvince = calleeInfo.Pos.Province
		cdrPkt.CalleeCity = calleeInfo.Pos.City

		cdrStr, err := json.Marshal(&cdrPkt)
		if err != nil {
			log.Errorf("json.Marshal error: ", err)
			return
		}

		// 推送至诈骗分析模型
		consumer.Next.Log(k, string(cdrStr))

		// 插入话单数据库
		dao.LogCDR(cdrPkt)
	} else {
		log.Debug("no handler for else condition")
	}
}
