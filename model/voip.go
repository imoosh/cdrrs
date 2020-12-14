package model

import (
	cmap "centnet-cdrrs/adapter/cache"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/model/prot/sip"
	"encoding/json"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	InviteCallIdPrefix = iota
	ByeCallIdPrefix
)

var (
	sipPool               = sync.Pool{New: func() interface{} { return &AnalyticSipPacket{} }}
	CallIdCache           = cmap.NewExpireMap()
	errUnresolvableNumber = errors.New("unresolvable number")
	errResolveSipPacket   = errors.New("resolve sip packet error")
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

func cacheInviteCallId(id string) (interface{}, bool) {
	return CallIdCache.SetWithExpire(id, InviteCallIdPrefix)
}

func cacheByeCallId(id string) (interface{}, bool) {
	return CallIdCache.SetWithExpire(id, ByeCallIdPrefix)
}

func isCachedInviteCallId(id string) bool {
	_, ok := CallIdCache.Get(id)
	return ok
}

func isCachedByeCallId(id string) bool {
	_, ok := CallIdCache.Get(id)
	return ok
}

func deleteInviteCall(id string) {
	CallIdCache.Del(id)
}

func deleteByeCall(id string) {
	CallIdCache.Del(id)
}

func NewSip() *AnalyticSipPacket {
	return sipPool.Get().(*AnalyticSipPacket)
}

func (pkt *AnalyticSipPacket) Free() {
	sipPool.Put(pkt)
}

func (pkt *AnalyticSipPacket) Init(asp *AnalyticSipPacket) *AnalyticSipPacket {
	pkt.EventId = asp.EventId
	pkt.EventTime = asp.EventTime
	pkt.Sip = asp.Sip
	pkt.Sport = asp.Sport
	pkt.Dip = asp.Dip
	pkt.Dport = asp.Dport
	pkt.CallId = asp.CallId
	pkt.CseqMethod = asp.CseqMethod
	pkt.ReqStatusCode = asp.ReqStatusCode
	pkt.FromUser = asp.FromUser
	pkt.ToUser = asp.ToUser
	pkt.UserAgent = asp.UserAgent
	pkt.CalleeInfo = asp.CalleeInfo
	pkt.GetAgain = asp.GetAgain
	return pkt
}

func (pkt *AnalyticSipPacket) doInvite200OKMessage() error {
	// 尝试本地缓存，如果缓存失败，则删除缓存并返回已缓存的数据，即BYE-200OK CallId
	if v, ok := cacheInviteCallId(pkt.CallId); ok {
		// 序列化sip报文
		pktStr, err := json.Marshal(pkt)
		if err != nil {
			return err
		}

		// invite插入redis
		redis.AsyncStoreWithExpire(pkt.CallId, string(pktStr), time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
		return nil
	} else {
		if v.(int) == ByeCallIdPrefix {
			redis.AsyncLoad(pkt.CallId, redis.DelayHandleUnit{
				Func: cdrRestore,
				Args: pkt,
			})
			//redis.AsyncDelete(key)
			return nil
		}
	}

	return errResolveSipPacket
}

func (pkt *AnalyticSipPacket) doBye200OKMessage() error {
	// 尝试本地缓存，如果缓存失败，则删除缓存并返回已缓存的数据，即INVITE-200OK CallId
	if v, ok := cacheByeCallId(pkt.CallId); ok {
		// 序列化sip报文
		pktStr, err := json.Marshal(pkt)
		if err != nil {
			return err
		}

		// invite插入redis
		redis.AsyncStoreWithExpire(pkt.CallId, string(pktStr), time.Second*time.Duration(conf.Conf.Redis.CacheExpire))
		return nil
	} else {
		if v.(int) == InviteCallIdPrefix {
			redis.AsyncLoad(pkt.CallId, redis.DelayHandleUnit{
				Func: cdrRestore,
				Args: pkt,
			})
			//redis.AsyncDelete(key)
			return nil
		}
	}

	return errResolveSipPacket
}

func validatePhoneNumber(num string) bool {
	for _, v := range num {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

type CalleeInfo struct {
	Number string
	Pos    dao.PhonePosition
}

func (info *CalleeInfo) parse(num string) error {
	length := len(num)
	if !validatePhoneNumber(num) || length < 11 {
		return errUnresolvableNumber
	}

	var (
		err       error
		calleeNum = num
	)

	if length >= 11 {
		callee := calleeNum[length-11:]
		if strings.HasPrefix(callee, "1") {
			// 手机号码归属查询
			if info.Pos, err = dao.GetPositionByMobilePhoneNumber(callee); err == nil {
				return nil
			}
		} else if strings.HasPrefix(callee, "0") {
			// 座机号码归属查询
			if info.Pos, err = dao.GetPositionByFixedPhoneNumber(callee, -1); err == nil {
				return nil
			}
		}
	}

	if length >= 12 {
		callee := calleeNum[length-12:]
		if strings.HasPrefix(callee, "0") {
			// 座机号码归属查询
			if info.Pos, err = dao.GetPositionByFixedPhoneNumber(callee, 4); err == nil {
				return nil
			}
		} else if strings.HasPrefix(callee, "86") {
			// 86xx xxxx xxxx -> 80xx xxxx xxxx -> 0xx xxxx xxxx
			callee = strings.Replace(callee, "86", "80", 1)[1:]
			if info.Pos, err = dao.GetPositionByFixedPhoneNumber(callee, -1); err == nil {
				return nil
			}
		}
	}

	if length >= 13 {
		callee := calleeNum[length-13:]
		if strings.HasPrefix(callee, "86") {
			// 86xxx xxxx xxxx -> 80xxx xxxx xxxx -> 0xxx xxxx xxxx
			callee = strings.Replace(callee, "86", "80", 1)[1:]
			if info.Pos, err = dao.GetPositionByFixedPhoneNumber(callee, 4); err == nil {
				return nil
			}
		}
	}

	{
		callee := calleeNum[length-11:]
		if strings.HasPrefix(callee, "852") {
			info.Number = callee
			info.Pos = dao.PhonePosition{Province: "香港", City: "香港"}
			return nil
		} else if strings.HasPrefix(callee, "853") {
			info.Number = callee
			info.Pos = dao.PhonePosition{Province: "澳门", City: "澳门"}
			return nil
		}
	}
	return errUnresolvableNumber
}

func atoi(s string, n int) (int, error) {
	if len(s) == 0 {
		return n, nil
	}

	return strconv.Atoi(s)
}

func (pkt *AnalyticSipPacket) parseSipPacket(rtd *RawTextData) error {
	sipMsg := sip.Parse([]byte(rtd.ParamContent))
	//sip.PrintSipStruct(&sipMsg)

	pkt.EventId = rtd.EventId
	pkt.EventTime = rtd.EventTime
	pkt.Sip = rtd.SaddrV4
	pkt.Sport = 0
	pkt.Dip = rtd.DaddrV4
	pkt.Dport = 0
	pkt.CallId = string(sipMsg.CallId.Value)
	pkt.CseqMethod = string(sipMsg.Cseq.Method)
	pkt.ReqStatusCode = 0
	//pkt.ReqMethod = string(sipMsg.Req.Method)
	//pkt.ReqUser = string(sipMsg.Req.User)
	//pkt.ReqHost = string(sipMsg.Req.Host)
	//pkt.ReqPort = 0
	//pkt.FromName = string(sipMsg.From.Name)
	pkt.FromUser = string(sipMsg.From.User)
	//pkt.FromHost = string(sipMsg.From.Host)
	//pkt.FromPort = 0
	//pkt.ToName = string(sipMsg.To.Name)
	pkt.ToUser = string(sipMsg.To.User)
	//pkt.ToHost = string(sipMsg.To.Host)
	//pkt.ToPort = 0
	//pkt.ContactName = string(sipMsg.Contact.Name)
	//pkt.ContactUser = string(sipMsg.Contact.User)
	//pkt.ContactHost = string(sipMsg.Contact.Host)
	//pkt.ContactPort = 0
	pkt.UserAgent = string(sipMsg.Ua.Value)

	// 没有call-id、cseq.method、直接丢弃
	if len(pkt.CallId) == 0 || len(pkt.CseqMethod) == 0 || len(pkt.ToUser) == 0 {
		return errInvalidSipPacketType
	}

	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
	var err error
	err = pkt.CalleeInfo.parse(pkt.ToUser)
	if err != nil {
		return errInvalidSipPacket
	}

	if pkt.Sport, err = atoi(rtd.Sport, 0); err != nil {
		return errInvalidSipPacket
	}
	if pkt.Dport, err = atoi(rtd.Dport, 0); err != nil {
		return errInvalidSipPacket
	}
	if pkt.ReqStatusCode, err = atoi(string(sipMsg.Req.StatusCode), 0); err != nil {
		return errInvalidSipPacket
	}
	//if pkt.ReqPort, err = atoi(string(sipMsg.Req.Port), 5060); err != nil {
	//	return pkt, errInvalidSipPacket
	//}
	//if pkt.FromPort, err = atoi(string(sipMsg.From.Port), 5060); err != nil {
	//	return pkt, errInvalidSipPacket
	//}
	//if pkt.ToPort, err = atoi(string(sipMsg.To.Port), 5060); err != nil {
	//	return pkt, errInvalidSipPacket
	//}
	//if pkt.ContactPort, err = atoi(string(sipMsg.Contact.Port), 5060); err != nil {
	//	return pkt, errInvalidSipPacket
	//}

	return nil
}
