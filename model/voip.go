package model

import (
	cmap "centnet-cdrrs/adapter/cache"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

const (
	InviteCallIdPrefix = "i"
	ByeCallIdPrefix    = "b"
)

var (
	CallIdCache           = cmap.NewExpireMap()
	errUnresolvableNumber = errors.New("unresolvable number")
)

type AnalyticSipPacket struct {
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

//func putCallIdInCache(id string) (interface{}, bool) {
//	return CallIdCache.SetWithExpire(id, nil, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
//}

func putInviteCallIdInCache(id string) (interface{}, bool) {
	return CallIdCache.SetWithExpire(id, InviteCallIdPrefix, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
}

func putByeCallIdInCache(id string) (interface{}, bool) {
	return CallIdCache.SetWithExpire(id, ByeCallIdPrefix, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
}

func isInviteCallIdInCache(id string) bool {
	_, ok := CallIdCache.Get(InviteCallIdPrefix + id)
	return ok
}

func isByeCallIdInCache(id string) bool {
	_, ok := CallIdCache.Get(ByeCallIdPrefix + id)
	return ok
}

func deleteInviteCallIdCache(id string) {
	CallIdCache.Del(InviteCallIdPrefix + id)
}

func deleteByeCallIdCache(id string) {
	CallIdCache.Del(ByeCallIdPrefix + id)
}

func doInvite200OKMessage(pkt AnalyticSipPacket, key, value string) {

	if v, ok := putInviteCallIdInCache(key); ok {
		// invite插入redis
		redis.AsyncStoreWithExpire(key, value, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
	} else {
		if v.(string) == ByeCallIdPrefix {
			redis.AsyncLoad(key, redis.DelayHandleUnit{
				Func: cdrRestore,
				Args: pkt,
			})
		}
		// 获取到后，立即删除缓存
		redis.AsyncDelete(pkt.CallId)
	}
}

func doBye200OKMessage(pkt AnalyticSipPacket, key, value string) {
	if v, ok := putByeCallIdInCache(key); ok {
		// invite插入redis
		redis.AsyncStoreWithExpire(key, value, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
	} else {
		if v.(string) == InviteCallIdPrefix {
			redis.AsyncLoad(key, redis.DelayHandleUnit{
				Func: cdrRestore,
				Args: pkt,
			})
		}
		// 获取到后，立即删除缓存
		redis.AsyncDelete(pkt.CallId)
	}
}

func validatePhoneNumber(num string) bool {
	for _, v := range num {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func parseCalleeInfo(num string) (CalleeInfo, error) {
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

func atoi(s string, n int) (int, error) {
	if len(s) == 0 {
		return n, nil
	}

	return strconv.Atoi(s)
}

// 解析sip报文
//func AnalyzePacket(consumer *kafka.ConsumerGroupMember, key, value interface{}) {
//
//	consumer.TotalCount++
//	consumer.TotalBytes = consumer.TotalBytes + uint64(len(value.([]byte)))
//
//	rtd, err := parseRawData(string(value.([]byte)))
//	if err != nil {
//		return
//	}
//	sipMsg := sip.Parse([]byte(rtd.ParamContent))
//
//	//sip.PrintSipStruct(&sipMsg)
//
//	pkt := AnalyticSipPacket{
//		EventId:       rtd.EventId,
//		EventTime:     rtd.EventTime,
//		Sip:           rtd.SaddrV4,
//		Sport:         0,
//		Dip:           rtd.DaddrV4,
//		Dport:         0,
//		CallId:        string(sipMsg.CallId.Value),
//		CseqMethod:    string(sipMsg.Cseq.Method),
//		ReqMethod:     string(sipMsg.Req.Method),
//		ReqStatusCode: 0,
//		ReqUser:       string(sipMsg.Req.User),
//		ReqHost:       string(sipMsg.Req.Host),
//		ReqPort:       0,
//		FromName:      string(sipMsg.From.Name),
//		FromUser:      string(sipMsg.From.User),
//		FromHost:      string(sipMsg.From.Host),
//		FromPort:      0,
//		ToName:        string(sipMsg.To.Name),
//		ToUser:        string(sipMsg.To.User),
//		ToHost:        string(sipMsg.To.Host),
//		ToPort:        0,
//		ContactName:   string(sipMsg.Contact.Name),
//		ContactUser:   string(sipMsg.Contact.User),
//		ContactHost:   string(sipMsg.Contact.Host),
//		ContactPort:   0,
//		UserAgent:     string(sipMsg.Ua.Value),
//	}
//
//	// 没有call-id、cseq.method、直接丢弃
//	if len(pkt.CallId) == 0 || len(pkt.CseqMethod) == 0 || len(pkt.ToUser) == 0 {
//		return
//	}
//
//	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
//	pkt.CalleeInfo, err = parseCalleeInfo(pkt.ToUser)
//	if err != nil {
//		log.Debug("DIRTY-DATA:", string(value.([]byte)))
//		return
//	}
//
//	if pkt.Sport, err = atoi(rtd.Sport, 0); err != nil {
//		return
//	}
//	if pkt.Dport, err = atoi(rtd.Dport, 0); err != nil {
//		return
//	}
//	if pkt.ReqStatusCode, err = atoi(string(sipMsg.Req.StatusCode), 0); err != nil {
//		return
//	}
//	if pkt.ReqPort, err = atoi(string(sipMsg.Req.Port), 5060); err != nil {
//		return
//	}
//	if pkt.FromPort, err = atoi(string(sipMsg.From.Port), 5060); err != nil {
//		return
//	}
//	if pkt.ToPort, err = atoi(string(sipMsg.To.Port), 5060); err != nil {
//		return
//	}
//	if pkt.ContactPort, err = atoi(string(sipMsg.Contact.Port), 5060); err != nil {
//		return
//	}
//
//	jsonStr, err := json.Marshal(pkt)
//	if err != nil {
//		log.Error(err)
//		return
//	}
//
//	consumer.Next.Log(string(sipMsg.CallId.Value), string(jsonStr))
//}
