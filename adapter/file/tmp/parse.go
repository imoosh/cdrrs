package tmp

import (
	"centnet-cdrrs/library/log"
	"errors"
	"strings"
)

var (
	ErrInvalidRawTextData   = errors.New("split error")
	ErrInvalidSipPacketType = errors.New("not invite/bye 200 ok message")
)

const SEPARATOR = "\r\n"

// 数据格式: "event_time", "event_id", saddr_v4", "daddr_v4", "dport", "param_content"，每条数据以"\r\n"分割

type RawTextData struct {
	EventTime    string `json:"event_time"`
	EventId      string `json:"event_id"`
	SaddrV4      string `json:"saddr_v4"`
	Sport        string `json:"sport"`
	DaddrV4      string `json:"daddr_v4"`
	Dport        string `json:"dport"`
	ParamContent string `json:"param_content"`
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

func ParseLine(s string) (rtd RawTextData, err error) {
	ss := split(pretreatment(s))
	if len(ss) == 0 {
		log.Errorf("split error: %s", s)
		err = ErrInvalidRawTextData
		return
	}

	// 只关注INVITE 200OK消息 和 BYE 200OK消息
	if !strings.HasPrefix(ss[6], "SIP/2.0 200 OK") {
		err = ErrInvalidSipPacketType
		return
	}

	rtd.EventTime = ss[0]
	rtd.EventId = ss[1]
	rtd.SaddrV4 = ss[2]
	rtd.Sport = ss[3]
	rtd.DaddrV4 = ss[4]
	rtd.Dport = ss[5]
	rtd.ParamContent = ss[6]

	return
}
