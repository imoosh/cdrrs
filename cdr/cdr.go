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
	Duration       string
}

func init() {
	dbLink := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?charset=%s&loc=%v", USERNAME, PASSWORD,
		NETWORK, SERVER, PORT, DATABASE, CHARSET)

	err := orm.RegisterDataBase("default", "mysql", dbLink)
	if err != nil {
		fmt.Println(err)
	}
}

func ParseInvite(invite []byte) {

	var invitedata Sip
	err := json.Unmarshal(invite, &invitedata)
	if err != nil {
		fmt.Println(err)
	}

	var byedata Sip
	err = json.Unmarshal(invite, &byedata)
	if err != nil {
		fmt.Println(err)
	}

	o := orm.NewOrm()
	o.Insert(&invitedata)
}

func ParseBye(bye []byte) {
	var byedata, invitedata Sip

	err := json.Unmarshal(bye, &byedata)
	if err != nil {
		fmt.Println(err)
	}
	qs := orm.NewOrm().QueryTable("sip").Filter("call_id", byedata.CallId).One(&invitedata)
	var cdrdata Cdr
	cdrdata.CallId = invitedata.CallId
	cdrdata.CallerIp = invitedata.Dip
	cdrdata.CallerPort = invitedata.Dport
	cdrdata.CalleeIp = invitedata.Sip
	cdrdata.CalleePort = invitedata.Sport
	cdrdata.CallerNum = invitedata.FromUser
	cdrdata.CalleeNum = invitedata.ToUser
	cdrdata.CallerDevice = byedata.CallId
	cdrdata.CalleeDevice = invitedata.UserAgent
	cdrdata.CalleeProvince = byedata.CallId
	cdrdata.CalleeCity = byedata.CallId
	cdrdata.ConnectTime = byedata.CallId
	cdrdata.DisconnectTime = byedata.CallId
	cdrdata.Duration = byedata.CallId
	//generateCdr(qs)
	orm.NewOrm().Insert(&byedata)
}

func generateCdr(qs *orm.querySet) {
	selectsql := `
	SELECT
	call_id,
	caller_ip,
	caller_port,
	callee_ip,
	callee_port,
	caller_number,
	callee_num,
	caller_device,
	callee_device,
	province AS callee_province,
	city AS callee_city,
	connect_time,
	disconnect_time,
	duration
FROM
	(
		SELECT
			call_id,
			sip AS caller_ip,
			sport AS caller_port,
			dip AS callee_ip,
			dport AS callee_port,
			from_user AS caller_number,
			to_user AS callee_num,
			GROUP_CONCAT(
				CASE
				WHEN cseq_method = "BYE" && req_status_code = "200" THEN
					user_agent
				END
			) AS caller_device,
			GROUP_CONCAT(
				CASE
				WHEN cseq_method = "INVITE" && req_status_code = "200" THEN
					user_agent
				END
			) AS callee_device,
			GROUP_CONCAT(
				CASE
				WHEN cseq_method = "INVITE" && req_status_code = "200" THEN
					event_time
				END
			) AS connect_time,
			GROUP_CONCAT(
				CASE
				WHEN cseq_method = "BYE" && req_status_code = "200" THEN
					event_time
				END
			) AS disconnect_time,
			TIMESTAMPDIFF(
				SECOND,
				GROUP_CONCAT(
					CASE
					WHEN cseq_method = "INVITE" && req_status_code = "200" THEN
						event_time
					END
				),
				GROUP_CONCAT(
					CASE
					WHEN cseq_method = "BYE" && req_status_code = "200" THEN
						event_time
					END
				)
			) AS duration
		FROM
			sip_analytic_packet
		GROUP BY
			call_id
	) AS a
LEFT JOIN phone_position b ON SUBSTR(a.callee_num, 2, 7) = b.phone
	`

	insertsql := `
	INSERT INTO cdr (
		call_id,
		caller_ip,
		caller_port,
		callee_ip,
		callee_port,
		caller_num,
		callee_num,
		caller_device,
		callee_device,
		connect_time,
		disconnect_time,
		duration
	)` + selectsql

	_, err := orm.NewOrm().Raw(insertsql).Exec()
	if err != nil {
		fmt.Println(err)
	}
}
