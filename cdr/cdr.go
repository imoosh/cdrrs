package cdr

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
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

type Sip struct {
	EventId       string `json:"event_id" orm:"column(event_id)"`
	EventTime     string `json:"event_time" orm:"column(event_time)"`
	Sip           string `json:"sip" orm:"column(sip)"`
	Sport         int    `json:"sport" orm:"column(sport)"`
	Dip           string `json:"dip" orm:"column(dip)"`
	Dport         int    `json:"dport" orm:"column(dport)"`
	CallId        string `json:"call_id" orm:"column(call_id)"`
	CseqMethod    string `json:"cseq_method" orm:"column(cseq_method)"`
	ReqMethod     string `json:"req_method" orm:"column(req_method)"`
	ReqStatusCode string `json:"req_status_code" orm:"column(req_status_code)"`
	ReqUser       string `json:"req_user" orm:"column(req_user)"`
	ReqHost       string `json:"req_host" orm:"column(req_host)"`
	ReqPort       int    `json:"req_port" orm:"column(req_port)"`
	FromName      string `json:"from_name" orm:"column(from_name)"`
	FromUser      string `json:"from_user" orm:"column(from_user)"`
	FromHost      string `json:"from_host" orm:"column(from_host)"`
	FromPort      int    `json:"from_port" orm:"column(from_port)"`
	ToName        string `json:"to_name" orm:"column(to_name)"`
	ToUser        string `json:"to_user" orm:"column(to_user)"`
	ToHost        string `json:"to_host" orm:"column(to_host)"`
	ToPort        int    `json:"to_port" orm:"column(to_port)"`
	ContactName   string `json:"contact_name" orm:"column(contact_name)"`
	ContactUser   string `json:"contact_user" orm:"column(contact_user)"`
	ContactHost   string `json:"contact_host" orm:"column(contact_host)"`
	ContactPort   string `json:"contact_port" orm:"column(contact_port)"`
	UserAgent     string `json:"user_agent" orm:"column(user_agent)"`
}

//type Cdr struct {
//	CallId         string `json:"call_id" orm:"column(call_id)"`
//	CallerIp       string `json:"caller_ip" orm:"column(caller_ip)"`
//	CallerPort     int    `json:"caller_ip" orm:"column(caller_port)"`
//	CalleeIp       string `json:"callee_ip" orm:"column(callee_ip)"`
//	CalleePort     int    `json:"callee_port" orm:"column(callee_port)"`
//	CallerNum      string `json:"caller_num" orm:"column(caller_num)"`
//	CalleeNum      string `json:"callee_num" orm:"column(callee_num)"`
//	CallerDevice   string `json:"callee_device" orm:"column(callee_ip)"`
//	CalleeDevice   string `json:"callee_device" orm:"column(callee_device)"`
//	CalleeProvince string `json:"callee_province" orm:"column(callee_province)"`
//	CalleeCity     string `json:"callee_city" orm:"column(callee_city)"`
//	ConnectTime    string `json:"connect_time" orm:"column(connect_time)"`
//	DisconnectTime string `json:"disconnect_time" orm:"column(disconnect_time)"`
//	Duration       string `json:"duration" orm:"column(duration)"`
//}

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

type Attribution struct {
	province string
	city     string
}

//func init() {
//	dbLink := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", USERNAME, PASSWORD,
//		NETWORK, SERVER, PORT, DATABASE, CHARSET)
//
//	err := orm.RegisterDataBase("default", "mysql", dbLink)
//	if err != nil {
//		fmt.Println(err)
//	}
//}

func ParseInvite200OKMessage(invite []byte) {
	var inviteData Sip
	err := json.Unmarshal(invite, &inviteData)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(inviteData.CallId)
	if inviteData.CseqMethod == "INVITE" && inviteData.ReqStatusCode == "200" {
		o := orm.NewOrm()
		_, err := o.Insert(&inviteData)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func ParseBye(bye []byte) *Cdr {
	var bye200OKMsg, invite200OKMsg Sip
	var err error
	err = json.Unmarshal(bye, &bye200OKMsg)
	if err != nil {
		fmt.Println(err)
	}
	if bye200OKMsg.CseqMethod != "BYE" || bye200OKMsg.ReqStatusCode != "200" {
		return nil
	}
	//根据BYE 200OK包的call_id获取对应的INVITE 200OK的包
	err = orm.NewOrm().QueryTable("sip_analytic_packet").Filter("call_id", bye200OKMsg.CallId).One(&invite200OKMsg)
	jsonStr, _ := json.Marshal(invite200OKMsg)
	fmt.Println(jsonStr)
	//填充话单字段信息
	cdr := Cdr{
		CallId:         invite200OKMsg.CallId,
		CallerIp:       invite200OKMsg.Dip,
		CallerPort:     invite200OKMsg.Dport,
		CalleeIp:       invite200OKMsg.Sip,
		CalleePort:     invite200OKMsg.Sport,
		CallerNum:      invite200OKMsg.FromUser,
		CalleeDevice:   invite200OKMsg.UserAgent,
		CalleeNum:      invite200OKMsg.ToUser,
		ConnectTime:    invite200OKMsg.EventTime,
		DisconnectTime: bye200OKMsg.EventTime,
		Duration:       10,
	}
	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if invite200OKMsg.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent

	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}
	//获取被叫归属地
	cdr.CalleeProvince, cdr.CalleeCity = getCalleeAttribution(&invite200OKMsg)
	generateCdr(&cdr)
	_, err = orm.NewOrm().Insert(&bye200OKMsg)
	return &cdr
}

func getCalleeAttribution(inviteData *Sip) (province, city string) {
	sql := fmt.Sprintf("SELECT province,city FROM (SELECT to_user FROM sip_analytic_packet WHERE cseq_method = \"INVITE\" AND req_status_code = \"200\" AND call_id = %s) AS a LEFT JOIN phone_position b ON SUBSTR(a.to_user,2,7) = b.phone", inviteData.CallId)
	attribution := new(Attribution)
	err := orm.NewOrm().Raw(sql).QueryRow(&attribution)
	if err != nil {
		fmt.Println(err)
	}

	province = attribution.province
	city = attribution.city
	return
}

func generateCdr(cdr *Cdr) {
	sql := fmt.Sprintf("insert into voip_restored_cdr (%s,%s,%d,%s,%d,%s,%s,%s,%s,%s,%s,%s,%s,%d)", cdr.CallId, cdr.CallerIp, cdr.CallerPort,
		cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice, cdr.CalleeDevice, cdr.CalleeProvince,
		cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration)
	_, err := orm.NewOrm().Raw(sql).Exec()
	if err != nil {
		fmt.Println(err)
	}
}
