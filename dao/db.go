package dao

import (
	"centnet-cdrrs/library/log"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"time"
)

var phonePositionMap = make(map[string]interface{})

type Config struct {
	DSN string
}

type SipAnalyticPacket struct {
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
}

type VoipRestoredCdr struct {
	Id             int64     `json:"id"`
	CallId         string    `json:"callId"`
	CallerIp       string    `json:"callerIp"`
	CallerPort     int       `json:"callerPort"`
	CalleeIp       string    `json:"calleeIp"`
	CalleePort     int       `json:"calleePort"`
	CallerNum      string    `json:"callerNum"`
	CalleeNum      string    `json:"calleeNum"`
	CallerDevice   string    `json:"callerDevice"`
	CalleeDevice   string    `json:"calleeDevice"`
	CalleeProvince string    `json:"calleeProvince"`
	CalleeCity     string    `json:"calleeCity"`
	ConnectTime    time.Time `json:"connectTime"`
	DisconnectTime time.Time `json:"disconnectTime"`
	Duration       int       `json:"duration"`
	FraudType      string    `json:"fraudType"`
}

type PhonePosition struct {
	Id         int
	Prefix     string
	Phone      string
	Province   string
	ProvinceId int64
	City       string
	CityId     int64
	Isp        string
	Code1      string
	Zip        string
	Types      string
}

type DateTime time.Time

//func (t DateTime) MarshalJSON() ([]byte, error) {
//	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
//	return []byte(stamp), nil
//}

func (t DateTime) MarshalJSON() ([]byte, error) {
	return ([]byte)(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (cdr VoipRestoredCdr) MarshalJSON() ([]byte, error) {
	type TmpJSON VoipRestoredCdr
	return json.Marshal(&struct {
		TmpJSON
		ConnectTime    DateTime `json:"connectTime"`
		DisconnectTime DateTime `json:"disconnectTime"`
	}{
		TmpJSON:        (TmpJSON)(cdr),
		ConnectTime:    DateTime(cdr.ConnectTime),
		DisconnectTime: DateTime(cdr.DisconnectTime),
	})
}

func Init(c *Config) error {
	//orm.Debug = true
	err := orm.RegisterDataBase("default", "mysql", c.DSN, 30)
	if err != nil {
		log.Error(err)
		return errors.New("orm.RegisterDataBase failed")
	}

	orm.RegisterModel(new(SipAnalyticPacket))
	orm.RegisterModel(new(VoipRestoredCdr))
	orm.RegisterModel(new(PhonePosition))

	// 将号码归属地表预读到内存中，加速查询速度
	QueryPhonePosition()
	return nil
}
