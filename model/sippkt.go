package model

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
