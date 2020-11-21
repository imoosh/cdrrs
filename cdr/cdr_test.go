package cdr

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"testing"
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

func init() {
	dbLink := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", USERNAME, PASSWORD,
		NETWORK, SERVER, PORT, DATABASE, CHARSET)

	err := orm.RegisterDataBase("default", "mysql", dbLink)
	if err != nil {
		fmt.Println(err)
	}
}

func TestParseInvite(t *testing.T) {
	inviteok := []byte(`
	{
		"event_id": "10020044170",
  		"event_time": "20201111100820",
  		"sip": "219.143.187.139",
  		"sport": "5088",
  		"dip": "192.168.6.24",
  		"dport": "5060",
  		"call_id": "04ab01e2d142787@192.168.6.24",
	    "cseq_method": "INVITE",
	    "req_method": "",
	    "req_status_code": "200",
	    "req_user": "018926798345",
	    "req_host": "219.143.187.139",
	    "req_port": "5060",
	    "from_name": "1101385",
	    "from_user": "1101385",
	    "from_host": "192.168.6.24",
	    "from_port": "5060",
	    "to_name": "",
	    "to_user": "018926798345",
	    "to_host": "219.143.187.139",
	    "to_port": "5060",
	    "contact_name": "1101385",
	    "contact_user": "1101385",
	    "contact_host": "192.168.6.24",
	    "contact_port": "5060",
	    "user_agent": "DonJin SIP Server 3.2.0_i"
	}
	`)

	ParseInvite200OKMessage(inviteok)
}

func TestParseBye(t *testing.T) {
	byeok := []byte(`
	{
		"event_id": "10020044170",
	  	"event_time": "20201111100903",
	  	"sip": "219.143.187.139",
	    "sport": "5088",
	    "dip": "192.168.6.24",
	    "dport": "5060",
	    "call_id": "04ab01e2d142787@192.168.6.24",
	    "cseq_method": "BYE",
	    "req_method": "",
	    "req_status_code": "200",
	    "req_user": "",
	    "req_host": "",
	    "req_port": "5060",
	    "from_name": "1101385",
	    "from_user": "1101385",
	    "from_host": "192.168.6.24",
	    "from_port": "5060",
	    "to_name": "",
	    "to_user": "018926798345",
	    "to_host": "219.143.187.139",
	    "to_port": "5060",
	    "contact_name": "1101385",
	    "contact_user": "1101385",
	    "contact_host": "192.168.6.24",
	    "contact_port": "5060",
	    "user_agent": ""
	}
	`)

	ParseBye(byeok)
}
