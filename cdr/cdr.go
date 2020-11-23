package cdr

import (
	_ "github.com/go-sql-driver/mysql"
	//"strconv"
)

//const (
//	USERNAME = "root"
//	PASSWORD = "123456"
//	NETWORK  = "tcp"
//	SERVER   = "192.168.1.205"
//	PORT     = 3306
//	DATABASE = "centnet_voip"
//	CHARSET  = "utf8"
//)

type Cdr struct {
	CallId         string
	CallerIp       string
	CallerPort     int
	CalleeIp       string
	CalleePort     int
	CallerNum      string
	CalleeNum      string
	CallerDevice   string
	CalleeDevice   string
	CalleeProvince string
	CalleeCity     string
	ConnectTime    string
	DisconnectTime string
	Duration       int
}
