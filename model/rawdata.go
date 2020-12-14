package model

import (
	"centnet-cdrrs/library/log"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

const SEPARATOR = "\r\n"

var (
	rtdPool                 = sync.Pool{New: func() interface{} { return &RawTextData{} }}
	errInvalidRawTextData   = errors.New("split error")
	errInvalidSipPacketType = errors.New("not invite/bye 200 ok message")
	errInvalidSipPacket     = errors.New("not invite/bye 200 ok message")
)

// 数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每条数据以"\r\n"分割

type RawTextData struct {
	EventTime    string
	EventId      string
	SaddrV4      string
	Sport        string
	DaddrV4      string
	Dport        string
	ParamContent string
}

func NewRtd() *RawTextData {
	return rtdPool.Get().(*RawTextData)
}

func (rtd *RawTextData) Free() {
	rtdPool.Put(rtd)
}

// 根据安全中心提供的数据样例，替换特殊字符
func pretreatment(s string) string {
	s = strings.ReplaceAll(s, "\\0D\\0A", "\r\n")
	return strings.ReplaceAll(s, "\"\"", "\"")
}

// 数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每个字段间以","间隔
func split(s string) []string {
	// 分割成7端
	ss := strings.SplitN(s, "\",\"", 7)
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

func (rtd *RawTextData) parse(s string) error {
	ss := split(pretreatment(s))
	if len(ss) == 0 {
		log.Errorf("split error: %s", s)
		return errInvalidRawTextData
	}

	// 只关注INVITE 200OK消息 和 BYE 200OK消息
	if !strings.HasPrefix(ss[6], "SIP/2.0 200 OK") {
		return errInvalidSipPacketType
	}

	rtd.EventTime = ss[0]
	rtd.EventId = ss[1]
	rtd.SaddrV4 = ss[2]
	rtd.Sport = ss[3]
	rtd.DaddrV4 = ss[4]
	rtd.Dport = ss[5]
	rtd.ParamContent = ss[6]
	return nil
}
