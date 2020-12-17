package tmp

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

var sipPool = sync.Pool{New: func() interface{} { return new(AnalyticSipPacket) }}
var emptySip = AnalyticSipPacket{}

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

func (sip *AnalyticSipPacket) reset() {
	*sip = emptySip
}

type CalleeInfo struct {
	Num string
	Pos PhonePosition
}

type PhonePosition struct {
	Id         int    `json:"-"`
	Prefix     string `json:"-"`
	Phone      string `json:"-"`
	Province   string `json:"province"`
	ProvinceId int64  `json:"-"`
	City       string `json:"city"`
	CityId     int64  `json:"-"`
	Isp        string `json:"-"`
	Code1      string `json:"-"`
	Zip        string `json:"-"`
	Types      string `json:"-"`
}

func Local() AnalyticSipPacket {
	sip := AnalyticSipPacket{
		EventId:       uuid.NewV4().String(),
		EventTime:     uuid.NewV4().String(),
		Sip:           uuid.NewV4().String(),
		Sport:         2353,
		Dip:           uuid.NewV4().String(),
		Dport:         3432,
		CallId:        uuid.NewV4().String(),
		CseqMethod:    uuid.NewV4().String(),
		ReqStatusCode: 4363,
		FromUser:      uuid.NewV4().String(),
		ToUser:        uuid.NewV4().String(),
		UserAgent:     uuid.NewV4().String(),
		GetAgain:      false,
	}
	return sip
}

func New() *AnalyticSipPacket {
	sip := sipPool.Get().(*AnalyticSipPacket)
	sip.EventId = uuid.NewV4().String()
	sip.EventTime = uuid.NewV4().String()
	sip.Sip = uuid.NewV4().String()
	sip.Sport = 2353
	sip.Dip = uuid.NewV4().String()
	sip.Dport = 3432
	sip.CallId = uuid.NewV4().String()
	sip.CseqMethod = uuid.NewV4().String()
	sip.ReqStatusCode = 4363
	sip.FromUser = uuid.NewV4().String()
	sip.ToUser = uuid.NewV4().String()
	sip.UserAgent = uuid.NewV4().String()
	sip.GetAgain = false

	return sip
}

func Free(sip *AnalyticSipPacket) {
	sipPool.Put(sip)
}
