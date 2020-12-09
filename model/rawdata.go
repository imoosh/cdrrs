package model

import (
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model/prot/sip"
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
)

const SEPARATOR = "\r\n"

var (
	ErrInvalidRawTextData   = errors.New("split error")
	ErrInvalidSipPacketType = errors.New("not invite/bye 200 ok message")
	ErrInvalidSipPacket     = errors.New("not invite/bye 200 ok message")
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

func parseRawData(s string) (rtd RawTextData, err error) {
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

func (rtd RawTextData) parseSipPacket() (AnalyticSipPacket, error) {
	sipMsg := sip.Parse([]byte(rtd.ParamContent))
	//sip.PrintSipStruct(&sipMsg)

	pkt := AnalyticSipPacket{
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
		return pkt, ErrInvalidSipPacketType
	}

	// 被叫号码字段未解析出手机号码或坐席号码归属地，直接丢弃(同一会话中所有包的FROM字段或TO字段都一样)
	var err error
	pkt.CalleeInfo, err = parseCalleeInfo(pkt.ToUser)
	if err != nil {
		//log.Debug("DIRTY-DATA:", string(value.([]byte)))
		return pkt, ErrInvalidSipPacket
	}

	if pkt.Sport, err = atoi(rtd.Sport, 0); err != nil {
		return pkt, ErrInvalidSipPacket
	}
	if pkt.Dport, err = atoi(rtd.Dport, 0); err != nil {
		return pkt, ErrInvalidSipPacket
	}
	if pkt.ReqStatusCode, err = atoi(string(sipMsg.Req.StatusCode), 0); err != nil {
		return pkt, ErrInvalidSipPacket
	}
	if pkt.ReqPort, err = atoi(string(sipMsg.Req.Port), 5060); err != nil {
		return pkt, ErrInvalidSipPacket
	}
	if pkt.FromPort, err = atoi(string(sipMsg.From.Port), 5060); err != nil {
		return pkt, ErrInvalidSipPacket
	}
	if pkt.ToPort, err = atoi(string(sipMsg.To.Port), 5060); err != nil {
		return pkt, ErrInvalidSipPacket
	}
	if pkt.ContactPort, err = atoi(string(sipMsg.Contact.Port), 5060); err != nil {
		return pkt, ErrInvalidSipPacket
	}

	return pkt, nil
}

func DoLine(line string) {

	// 解析原始包各字段
	rtd, err := parseRawData(line)
	if err != nil {
		return
	}

	// 解析sip报文
	pkt, err := rtd.parseSipPacket()
	if err != nil {
		return
	}

	// 序列化sip报文
	jsonStr, err := json.Marshal(pkt)
	if err != nil {
		return
	}

	// call-id与sip报文原始数据
	k, v := pkt.CallId, string(jsonStr)
	if pkt.CseqMethod == "INVITE" && pkt.ReqStatusCode == 200 {
		doInvite200OKMessage(pkt, k, v)
	} else if pkt.CseqMethod == "BYE" && pkt.ReqStatusCode == 200 {
		doBye200OKMessage(pkt, k, v)
	} else {
		log.Debug("no handler for else condition")
	}
}
