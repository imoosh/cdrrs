package cdr

import (
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"time"

	"encoding/json"
	"fmt"
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

func ParseInvite200OKMessage(key, value interface{}) {
	var invite200OKMsg dao.SipAnalyticPacket
	err := json.Unmarshal([]byte(value.(string)), &invite200OKMsg)
	if err != nil {
		fmt.Println(err)
	}
	if invite200OKMsg.CseqMethod == "INVITE" && invite200OKMsg.ReqStatusCode == 200 {
		//	invite200OK插入redis
		//redis.RedisConn.PutWithExpire(key.(string), value.(string), 5)
		redis.RedisConn.Put(key.(string), value.(string))
		fmt.Println("Insert redis ok")
	}
}

func ParseBye200OKMsg(key, value interface{}) *dao.VoipRestoredCdr {
	var bye200OKMsg, invite200OKMsg dao.SipAnalyticPacket
	var calleeAtrribution dao.PhonePosition
	var err error
	err = json.Unmarshal([]byte(value.(string)), &bye200OKMsg)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(bye200OKMsg.CseqMethod, bye200OKMsg.ReqStatusCode)
	if bye200OKMsg.CseqMethod != "BYE" || bye200OKMsg.ReqStatusCode != 200 {
		return nil
	}
	//根据BYE 200OK包的call_id从redis获取对应的INVITE 200OK的包
	redisInvite200OK, notGet := redis.RedisConn.Get(key.(string))
	if notGet != nil {
		fmt.Println("get invite200ok failed")
		return nil
	}
	err = json.Unmarshal([]byte(redisInvite200OK), &invite200OKMsg)
	if err != nil {
		fmt.Println(err)
	}
	//查询成功后删除redis中INVITE 200OK包
	redis.RedisConn.Delete(key.(string))

	//填充话单字段信息
	cdr := dao.VoipRestoredCdr{
		CallId:       invite200OKMsg.CallId,
		CallerIp:     invite200OKMsg.Dip,
		CallerPort:   invite200OKMsg.Dport,
		CalleeIp:     invite200OKMsg.Sip,
		CalleePort:   invite200OKMsg.Sport,
		CallerNum:    invite200OKMsg.FromUser,
		CalleeNum:    invite200OKMsg.ToUser,
		CalleeDevice: invite200OKMsg.UserAgent,
	}
	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	if invite200OKMsg.Dip == bye200OKMsg.Sip {
		cdr.CallerDevice = bye200OKMsg.UserAgent
	} else {
		//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
		cdr.CallerDevice = ""
	}
	//获取被叫归属地
	calleeAtrribution = dao.GetPositionByPhoneNum(invite200OKMsg.ToUser)
	fmt.Println(invite200OKMsg.ToUser, calleeAtrribution.Province, calleeAtrribution.City)
	cdr.CalleeProvince = calleeAtrribution.Province
	cdr.CalleeCity = calleeAtrribution.City
	timeConnect := fmt.Sprintf("%s-%s-%s %s:%s:%s", invite200OKMsg.EventTime[0:4], invite200OKMsg.EventTime[4:6],
		invite200OKMsg.EventTime[6:8], invite200OKMsg.EventTime[8:10], invite200OKMsg.EventTime[10:12], invite200OKMsg.EventTime[12:])
	connectTimeStamp, _ := time.ParseInLocation("2006-01-02 15:04:05", timeConnect, time.Local)
	cdr.ConnectTime = timeConnect
	//localConnect = time.Unix(timeConnect, 0).Format("2006-01-02 15:04:05")
	timeDisconnect := fmt.Sprintf("%s-%s-%s %s:%s:%s", bye200OKMsg.EventTime[0:4], bye200OKMsg.EventTime[4:6],
		bye200OKMsg.EventTime[6:8], bye200OKMsg.EventTime[8:10], bye200OKMsg.EventTime[10:12], bye200OKMsg.EventTime[12:])
	fmt.Println(timeConnect, timeDisconnect)
	disconnectTimeStamp, _ := time.ParseInLocation("2006-01-02 15:04:05", timeDisconnect, time.Local)
	cdr.DisconnectTime = timeDisconnect
	cdr.Duration = int(disconnectTimeStamp.Sub(connectTimeStamp).Seconds())
	cdr.FraudType = ""
	fmt.Println("time:", invite200OKMsg.EventTime, bye200OKMsg.EventTime, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration)

	jsonStr, _ := json.Marshal(&cdr)
	fmt.Println(string(jsonStr))
	dao.InsertCDR(&cdr)
	fmt.Println("insert success")

	return &cdr

}
