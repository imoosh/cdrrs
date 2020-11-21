package cdr

import (
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const (
	USERNAME = "root"
	PASSWORD = "123456"
	NETWORK  = "tcp"
	SERVER   = "192.168.1.205"
	PORT     = 3306
	DATABASE = "centnet_voip"
	CHARSET  = "utf8"
)

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

func init() {
	dbLink := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", USERNAME, PASSWORD,
		NETWORK, SERVER, PORT, DATABASE, CHARSET)

	err := orm.RegisterDataBase("default", "mysql", dbLink)
	if err != nil {
		log.Error(err)
	}
}

func ParseInvite200OKMessage(invite []byte) {
	var invite200OKMsg dao.SipAnalyticPacket
	err := json.Unmarshal(invite, &invite200OKMsg)
	if err != nil {
		log.Error(err)
	}

	log.Debug(invite200OKMsg.CallId)
	if invite200OKMsg.CseqMethod == "INVITE" && invite200OKMsg.ReqStatusCode == 200 {
		o := orm.NewOrm()
		_, err := o.Insert(&invite200OKMsg)
		if err != nil {
			log.Error(err)
		}
	}
}

func ParseBye200OKMsg(bye []byte) *dao.VoipRestoredCdr {
	var bye200OKMsg, invite200OKMsg dao.SipAnalyticPacket
	var err error
	err = json.Unmarshal(bye, &bye200OKMsg)
	if err != nil {
		log.Error(err)
	}
	if bye200OKMsg.CseqMethod != "BYE" || bye200OKMsg.ReqStatusCode != 200 {
		return nil
	}
	//根据BYE 200OK包的call_id从数据库获取对应的INVITE 200OK的包
	dao.GetInvite200OKMsg(&bye200OKMsg)
	jsonStr, _ := json.Marshal(invite200OKMsg)
	log.Debug(jsonStr)
	//填充话单字段信息
	cdr := dao.VoipRestoredCdr{
		CallId:       invite200OKMsg.CallId,
		CallerIp:     invite200OKMsg.Dip,
		CallerPort:   invite200OKMsg.Dport,
		CalleeIp:     invite200OKMsg.Sip,
		CalleePort:   invite200OKMsg.Sport,
		CallerNum:    invite200OKMsg.FromUser,
		CalleeDevice: invite200OKMsg.UserAgent,
		CalleeNum:    invite200OKMsg.ToUser,
		//ConnectTime:    invite200OKMsg.EventTime,
		//DisconnectTime: bye200OKMsg.EventTime,
		Duration: 10,
	}
	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if invite200OKMsg.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent
	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}
	//获取被叫归属地
	cdr.CalleeProvince, cdr.CalleeCity = dao.GetCalleeAttribution(&invite200OKMsg)
	cdr.ConnectTime, _ = time.Parse("2006-01-02 15:04:05", invite200OKMsg.EventTime)
	cdr.DisconnectTime, _ = time.Parse("2006-01-02 15:04:05", bye200OKMsg.EventTime)
	cdr.Duration = cdr.ConnectTime.Sub(cdr.DisconnectTime)
	dao.InsertCDR(&cdr)
	_, err = orm.NewOrm().Insert(&bye200OKMsg)
	if err != nil {
		log.Error(err)
	}

	return &cdr
}
