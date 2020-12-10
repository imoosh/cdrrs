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
	InviteCallIdPrefix = iota
	ByeCallIdPrefix
)

var (
	CallIdCache           = cmap.NewExpireMap()
	errUnresolvableNumber = errors.New("unresolvable number")
)

type AnalyticSipPacket struct {
	//Id            uint64 `json:"id"`
	EventId       string `json:"eventId"`
	EventTime     string `json:"eventTime"`
	Sip           string `json:"sip"`
	Sport         int    `json:"sport"`
	Dip           string `json:"dip"`
	Dport         int    `json:"dport"`
	CallId        string `json:"callId"`
	CseqMethod    string `json:"cseqMethod"`
	ReqStatusCode int    `json:"reqStatusCode"`
	//ReqMethod     string `json:"reqMethod"`
	//ReqUser       string `json:"reqUser"`
	//ReqHost       string `json:"reqHost"`
	//ReqPort       int    `json:"reqPort"`
	//FromName      string `json:"fromName"`
	FromUser string `json:"fromUser"`
	//FromHost      string `json:"fromHost"`
	//FromPort      int    `json:"fromPort"`
	//ToName        string `json:"toName"`
	ToUser string `json:"toUser"`
	//ToHost        string `json:"toHost"`
	//ToPort        int    `json:"toPort"`
	//ContactName   string `json:"contactName"`
	//ContactUser   string `json:"contactUser"`
	//ContactHost   string `json:"contactHost"`
	//ContactPort   int    `json:"contactPort"`
	UserAgent string `json:"userAgent"`

	CalleeInfo CalleeInfo `json:"calleeInfo"`
	GetAgain   bool       `json:"getAgain"`
}

type CalleeInfo struct {
	Number string
	Pos    dao.PhonePosition
}

//func putCallIdInCache(id string) (interface{}, bool) {
//	return CallIdCache.SetWithExpire(id, nil, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
//}

func putInviteCallIdInCache(id string) (interface{}, bool) {
	return CallIdCache.SetWithExpire(id, InviteCallIdPrefix)
}

func putByeCallIdInCache(id string) (interface{}, bool) {
	return CallIdCache.SetWithExpire(id, ByeCallIdPrefix)
}

func isInviteCallIdInCache(id string) bool {
	_, ok := CallIdCache.Get(id)
	return ok
}

func isByeCallIdInCache(id string) bool {
	_, ok := CallIdCache.Get(id)
	return ok
}

func deleteInviteCallIdCache(id string) {
	CallIdCache.Del(id)
}

func deleteByeCallIdCache(id string) {
	CallIdCache.Del(id)
}

func doInvite200OKMessage(pkt AnalyticSipPacket, key, value string) {
	if v, ok := putInviteCallIdInCache(key); ok {
		// invite插入redis
		redis.AsyncStoreWithExpire(key, value, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
	} else {
		if v.(int) == ByeCallIdPrefix {
			redis.AsyncLoad(key, redis.DelayHandleUnit{
				Func: cdrRestore,
				Args: pkt,
			})
			//redis.AsyncDelete(key)
		}
	}
}

func doBye200OKMessage(pkt AnalyticSipPacket, key, value string) {
	if v, ok := putByeCallIdInCache(key); ok {
		// invite插入redis
		redis.AsyncStoreWithExpire(key, value, time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
	} else {
		if v.(int) == InviteCallIdPrefix {
			redis.AsyncLoad(key, redis.DelayHandleUnit{
				Func: cdrRestore,
				Args: pkt,
			})
			//redis.AsyncDelete(key)
		}
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
