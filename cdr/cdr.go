package cdr

import (
	//"centnet-cdrrs/adapter/kafka/file"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	//"centnet-cdrrs/prot/sip"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	//"strconv"
	//"time"
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

//func init() {
//	dbLink := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=%s", USERNAME, PASSWORD,
//		NETWORK, SERVER, PORT, DATABASE, CHARSET)
//
//	err := orm.RegisterDataBase("default", "mysql", dbLink)
//	if err != nil {
//		log.Error(err)
//	}
//}

func ParseInvite200OKMessage(key, value []byte) {
	var invite200OKMsg dao.SipAnalyticPacket
	err := json.Unmarshal(value, &invite200OKMsg)
	if err != nil {
		log.Error(err)
	}

	if invite200OKMsg.CseqMethod == "INVITE" && invite200OKMsg.ReqStatusCode == 200 {
		//	invite200OK插入redis
		redis.RedisConn.Put(string(key), string(value))
		fmt.Println("Insert redis ok")
	}
}

func ParseBye200OKMsg(key, value []byte) *dao.VoipRestoredCdr {
	var bye200OKMsg, invite200OKMsg dao.SipAnalyticPacket
	var err error
	err = json.Unmarshal(value, &bye200OKMsg)
	if err != nil {
		log.Error(err)
	}
	if bye200OKMsg.CseqMethod != "BYE" || bye200OKMsg.ReqStatusCode != 200 {
		return nil
	}
	//根据BYE 200OK包的call_id从redis获取对应的INVITE 200OK的包
	redisInvite200OK, _ := redis.RedisConn.Get(string(key))
	fmt.Println(redisInvite200OK)
	err = json.Unmarshal([]byte(redisInvite200OK), &invite200OKMsg)
	if err != nil {
		fmt.Println(err)
	}
	//rtd := file.Parse(redisInvite200OK)
	//invite200OKMsg := sip.Parse([]byte(rtd.ParamContent))
	//jsonStr, _ := json.Marshal(invite200OKMsg)
	//log.Debug(jsonStr)

	//填充话单字段信息
	//cdr := dao.VoipRestoredCdr{
	//	CallId:       string(invite200OKMsg.CallId.Value),
	//	CallerIp:     rtd.DaddrV4,
	//	CallerPort:   0,
	//	CalleeIp:     rtd.SaddrV4,
	//	CalleePort:   0,
	//	CallerNum:    string(invite200OKMsg.From.Name),
	//	CalleeNum:    string(invite200OKMsg.To.User),
	//	CalleeDevice: string(invite200OKMsg.Ua.Value),
	//
	//	//ConnectTime:    invite200OKMsg.EventTime,
	//	//DisconnectTime: bye200OKMsg.EventTime,
	//	Duration: 10,
	//}

	//if cdr.CallerPort, err = strconv.Atoi(rtd.Dport); err != nil {
	//	log.Errorf("cannot convert %s to an integer", rtd.Sport)
	//	return nil
	//}
	//if cdr.CalleePort, err = strconv.Atoi(rtd.Sport); err != nil {
	//	log.Errorf("cannot convert %s to an integer", rtd.Sport)
	//	return nil
	//}
	//INVITE的200OK的目的ip为主叫ip，若与BYE的200OK源ip一致，则是被叫挂机，主叫发200 OK,此时user_agent为主叫设备信息
	//if invite200OKMsg.Dip == bye200OKMsg.Sip {
	//	cdr.CallerDevice = string(invite200OKMsg.Ua.Value)
	//} else {
	//	//主叫挂机，user_agent为被叫设备信息，此时获取不到主叫设备信息
	//	cdr.CallerDevice = ""
	//}
	//获取被叫归属地
	//cdr.CalleeProvince, cdr.CalleeCity = dao.GetCalleeAttribution(&invite200OKMsg)
	//cdr.ConnectTime, _ = time.Parse("2006-01-02 15:04:05", invite200OKMsg.EventTime)
	//cdr.DisconnectTime, _ = time.Parse("2006-01-02 15:04:05", bye200OKMsg.EventTime)
	//cdr.Duration = cdr.ConnectTime.Sub(cdr.DiscosnnectTime)
	//cdr.Duration = int(cdr.DisconnectTime.Sub(cdr.ConnectTime).Seconds())

	//dao.InsertCDR(&cdr)

	//return &cdr
	return nil
}
