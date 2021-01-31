package model

import (
	"bytes"
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/model/prot/sip"
	"errors"
	"strconv"
	"strings"
	"sync"
)

type SipPacket struct {
	//Id            uint64 `json:"id"`
	//EventId       string `json:"eventId"`
	EventTime     string `json:"t"`
	Sip           string `json:"si"`
	Sport         int    `json:"sp"`
	Dip           string `json:"di"`
	Dport         int    `json:"dp"`
	CallId        string `json:"-"`
	CseqMethod    string `json:"cm"`
	ReqStatusCode int    `json:"-"`
	//ReqMethod     string `json:"reqMethod"`
	//ReqUser       string `json:"reqUser"`
	//ReqHost       string `json:"reqHost"`
	//ReqPort       int    `json:"reqPort"`
	//FromName      string `json:"fromName"`
	FromUser string `json:"fu"`
	//FromHost      string `json:"fromHost"`
	//FromPort      int    `json:"fromPort"`
	//ToName        string `json:"toName"`
	ToUser string `json:"tu"`
	//ToHost        string `json:"toHost"`
	//ToPort        int    `json:"toPort"`
	//ContactName   string `json:"contactName"`
	//ContactUser   string `json:"contactUser"`
	//ContactHost   string `json:"contactHost"`
	//ContactPort   int    `json:"contactPort"`
	UserAgent string `json:"ua"`
}

const SEPARATOR = "\r\n"

var (
	delimiter             = []byte("\",\"")
	oldLineBreak          = []byte("\\0D\\0A")
	newLineBreak          = []byte("\r\n")
	oldDoubleQuotes       = []byte("\"\"")
	newDoubleQuotes       = []byte("\"")
	sipStatus200OKPrefix  = []byte("SIP/2.0 200 OK")
	emptyRtd              = RawTextData{}
	rtdPool               = sync.Pool{New: func() interface{} { return &RawTextData{} }}
	errInvalidRawTextData = errors.New("split error")
)

var (
	errInvalidSipPacket = errors.New("not invite/bye 200 ok message")
)

func ExtractLegalNumber(num string) string {
	length := len(num)
	if !validateNumberString(num) || length < 11 {
		return ""
	}

	var (
		err       error
		calleeNum = num
	)

	if length >= 11 {
		callee := calleeNum[length-11:]
		if strings.HasPrefix(callee, "1") {
			// 手机号码归属查询
			if _, err = QueryMobileNumberPosition(callee); err == nil {
				return callee
			}
		} else if strings.HasPrefix(callee, "0") {
			// 座机号码归属查询
			if _, err = QueryFixedNumberPosition(callee); err == nil {
				return callee
			}
		}
	}

	if length >= 12 {
		callee := calleeNum[length-12:]
		if strings.HasPrefix(callee, "0") {
			// 座机号码归属查询
			if _, err = QueryFixedNumberPosition(callee); err == nil {
				return callee
			}
		} else if strings.HasPrefix(callee, "86") {
			// 86xx xxxx xxxx -> 80xx xxxx xxxx -> 0xx xxxx xxxx
			callee = strings.Replace(callee, "86", "80", 1)[1:]
			if _, err = QueryFixedNumberPosition(callee); err == nil {
				return callee
			}
		}
	}

	if length >= 13 {
		callee := calleeNum[length-13:]
		if strings.HasPrefix(callee, "86") {
			// 86xxx xxxx xxxx -> 80xxx xxxx xxxx -> 0xxx xxxx xxxx
			callee = strings.Replace(callee, "86", "80", 1)[1:]
			if _, err = QueryFixedNumberPosition(callee); err == nil {
				return callee
			}
		}
	}

	callee := calleeNum[length-11:]
	if strings.HasPrefix(callee, "852") {
		return callee
	} else if strings.HasPrefix(callee, "853") {
		return callee
	}

	return ""
}

func atoi(s string, n int) (int, error) {
	if len(s) == 0 {
		return n, nil
	}

	return strconv.Atoi(s)
}

func ParseSipPacket(line interface{}, pkt *SipPacket) error {
	ss := split(pretreatment(line.([]byte)))
	if len(ss) == 0 {
		return errInvalidRawTextData
	}

	// 只关注INVITE 200OK消息 和 BYE 200OK消息
	if !bytes.HasPrefix(ss[6], sipStatus200OKPrefix) {
		return errInvalidSipPacket
	}

	// 数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每个字段间以","间隔
	var rtd RawTextData
	rtd.EventTime = ss[0]
	//rtd.EventId = ss[1]
	rtd.SaddrV4 = ss[2]
	rtd.Sport = ss[3]
	rtd.DaddrV4 = ss[4]
	rtd.Dport = ss[5]
	rtd.ParamContent = ss[6]

	sipMsg := sip.Parse(ss[6])
	//sip.PrintSipStruct(&sipMsg)

	//pkt.EventId = string(rtd.EventId)
	pkt.EventTime = string(rtd.EventTime)
	pkt.Sip = string(rtd.SaddrV4)
	pkt.Sport = 0
	pkt.Dip = string(rtd.DaddrV4)
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
		return errInvalidSipPacket
	}

	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
	var err error
	pkt.ToUser = ExtractLegalNumber(pkt.ToUser)
	if len(pkt.ToUser) == 0 {
		return errInvalidSipPacket
	}

	if pkt.Sport, err = atoi(string(rtd.Sport), 0); err != nil {
		return errInvalidSipPacket
	}
	if pkt.Dport, err = atoi(string(rtd.Dport), 0); err != nil {
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

// 数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每条数据以"\r\n"分割

type RawTextData struct {
	EventTime []byte
	//EventId      []byte
	SaddrV4      []byte
	Sport        []byte
	DaddrV4      []byte
	Dport        []byte
	ParamContent []byte
}

// 根据安全中心提供的数据样例，替换特殊字符
func pretreatment(s []byte) []byte {
	return bytes.ReplaceAll(bytes.ReplaceAll(s, oldLineBreak, newLineBreak), oldDoubleQuotes, newDoubleQuotes)
}

// 数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每个字段间以","间隔
func split(s []byte) [][]byte {
	// 分割成7端
	ss := bytes.SplitN(s, delimiter, 7)
	if len(ss) != 7 {
		return nil
	}

	for _, val := range ss {
		if len(val) == 0 {
			return nil
		}
	}

	// 去掉第一个字段的上引号和最后一个字段的下引号
	ss[0] = ss[0][1:]
	ss[6] = ss[6][:len(ss[6])-1]

	return ss
}

func (rtd *RawTextData) parse(s []byte) error {
	ss := split(pretreatment(s))
	if len(ss) == 0 {
		log.Errorf("split error: %s", s)
		return errInvalidRawTextData
	}

	// 只关注INVITE 200OK消息 和 BYE 200OK消息
	if !bytes.HasPrefix(ss[6], sipStatus200OKPrefix) {
		return errInvalidSipPacket
	}

	rtd.EventTime = ss[0]
	//rtd.EventId = ss[1]
	rtd.SaddrV4 = ss[2]
	rtd.Sport = ss[3]
	rtd.DaddrV4 = ss[4]
	rtd.Dport = ss[5]
	rtd.ParamContent = ss[6]
	return nil
}
