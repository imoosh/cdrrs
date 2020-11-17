package dao

import (
    "github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"
)

type Sip struct {
    Id            uint64
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

func init()  {
    //orm.Debug = true
    orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/centnet_voip?charset=utf8mb4", 30)
    orm.RegisterModel(new(Sip))
}
