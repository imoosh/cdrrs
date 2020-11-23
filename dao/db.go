package dao

import (
	"centnet-cdrrs/library/log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var phonePositionMap = make(map[string]interface{})

type Config struct {
	DSN string
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
	//ConnectTime    string `json:"connectTime"`
	//DisconnectTime string `json:"disconnectTime"`
	Duration  int    `json:"duration"`
	FraudType string `json:"fraudType"`
}

type DateTime time.Time

func (t DateTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

//func (t DateTime) MarshalJSON() ([]byte, error) {
//	return ([]byte)(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
//}
//

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

func Init(c *Config) error {
	//orm.Debug = true
	err := orm.RegisterDataBase("default", "mysql", c.DSN, 30)
	if err != nil {
		log.Error(err)
		return errors.New("orm.RegisterDataBase failed")
	}

	orm.RegisterModel(new(VoipRestoredCdr))
	orm.RegisterModel(new(PhonePosition))

	// 将号码归属地表预读到内存中，加速查询速度
	QueryPhonePosition()
	return nil
}
