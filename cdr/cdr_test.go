package cdr

import (
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

//func init() {
//	dbLink := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", USERNAME, PASSWORD,
//		NETWORK, SERVER, PORT, DATABASE, CHARSET)
//
//	err := orm.RegisterDataBase("default", "mysql", dbLink)
//	if err != nil {
//		fmt.Println(err)
//	}
//}

func TestParseInvite200OKMessage(t *testing.T) {
	key := `04ab01e2d142787@192.168.6.24`
	inviteok := `
	{
	"id":10000,
    "eventId":"10020044170",
    "eventTime":"20201111100820",
    "sip":"219.143.187.139",
    "sport":5088,
    "dip":"192.168.6.24",
    "dport":5060,
    "callId":"04ab01e2d142787@192.168.6.24",
    "cseqMethod":"INVITE",
    "reqMethod":" ",
    "reqStatusCode":200,
    "reqUser":"018926798345",
    "reqHost":"219.143.187.139",
    "reqPort":5060,
    "fromName":"1101385",
    "fromUser":"1101385",
    "fromHost":"192.168.6.24",
    "fromPort":5060,
    "toName":" ",
    "toUser":"018926798345",
    "toHost":"219.143.187.139",
    "toPort":5060,
    "contactName":"1101385",
    "contactUser":"1101385",
    "contactHost":"192.168.6.24",
    "contactPort":5060,
    "userAgent":"DonJin SIP Server 3.2.0_i"
	}
	`
	ParseInvite200OKMessage(key, inviteok)
}

func TestParseBye(t *testing.T) {
	key := `04ab01e2d142787@192.168.6.24`
	byeOk := `
	{
	"id":10000,
	"eventId": "10020044170",
	"eventTime": "20201111100903",
	"sip": "219.143.187.139",
	"sport": 5088,
	"dip": "192.168.6.24",
	"dport": 5060,
	"callId": "04ab01e2d142787@192.168.6.24",
	"cseqMethod": "BYE",
	"cseqMethod": "",
	    "reqStatusCode": 200,
	    "reqUser": "",
	    "reqHost": "",
	    "reqPort": 5060,
	    "fromName": "1101385",
	    "fromUser": "1101385",
	    "fromHost": "192.168.6.24",
	    "fromPort": 5060,
	    "toName": "",
	    "toUser": "018926798345",
	    "toHost": "219.143.187.139",
	    "toPort": "5060",
	    "contactName": "1101385",
	    "contactUser": "1101385",
	    "contactHost": "192.168.6.24",
	    "contactPort": "5060",
	    "userAgent": ""
	}
	`
	ParseBye200OKMsg(key, byeOk)
}
