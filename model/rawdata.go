package model

import (
	"bytes"
	"centnet-cdrrs/library/log"
	"github.com/pkg/errors"
	"sync"
)

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
	errInvalidSipPacket   = errors.New("not invite/bye 200 ok message")
)

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

func NewRtd() *RawTextData {
	return rtdPool.Get().(*RawTextData)
}

func (rtd *RawTextData) Free() {
	*rtd = emptyRtd
	rtdPool.Put(rtd)
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
