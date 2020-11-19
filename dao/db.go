package dao

import (
	"centnet-cdrrs/library/log"
	"errors"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DSN string
}

type Sip struct {
	Id            uint64
	EventId       string
	EventTime     string
	Sip           string
	Sport         uint16
	Dip           string
	Dport         uint16
	CallId        string
	ReqMethod     string
	ReqStatusCode int
	CseqMethod    string
	ReqUser       string
	ReqHost       string
	ReqPort       int
	FromName      string
	FromUser      string
	FromHost      string
	FromPort      int
	ToName        string
	ToUser        string
	ToHost        string
	ToPort        int
	ContactName   string
	ContactUser   string
	ContactHost   string
	ContactPort   int
	UserAgent     string
}

func Init(c *Config) error {
	//orm.Debug = true
	err := orm.RegisterDataBase("default", "mysql", c.DSN, 30)
	if err != nil {
		log.Error(err)
		return errors.New("orm.RegisterDataBase failed")
	}

	orm.RegisterModel(new(Sip))

	return nil
}
