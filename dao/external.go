package dao

import (
	"bytes"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"centnet-cdrrs/prot/udp"
	"fmt"
	"github.com/astaxie/beego/orm"
)

type UnpackedMessage struct {
	EventId   string
	EventTime string
	UDP       *udp.UdpMsg
	SIP       *sip.SipMsg
}

type Attribution struct {
	province string
	city     string
}

func QueryPhonePosition() {
	o := orm.NewOrm()
	pp := new(PhonePosition)
	var pps []PhonePosition

	n, err := o.QueryTable(pp).All(&pps)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("query phone_position %d rows", n)

	for _, val := range pps {
		phonePositionMap[val.Phone] = val
	}
}

func GetPositionByPhoneNum(phone string) PhonePosition {
	if len(phone) != 12 {
		log.Error("invalid phone number:", phone)
		return PhonePosition{}
	}

	if pp, ok := phonePositionMap[phone[1:8]]; ok {
		return pp.(PhonePosition)
	}
	return PhonePosition{}
}

func InsertCDR(cdr *VoipRestoredCdr) {
	sql := fmt.Sprintf("insert into voip_restored_cdr (call_id,caller_ip,caller_port,callee_ip,callee_port,caller_num,callee_num,caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration)values(\"%s\",\"%s\",%d,\"%s\",%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",%d)",
		cdr.CallId, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
		cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration)
	_, err := orm.NewOrm().Raw(sql).Exec()
	if err != nil {
		log.Error(err)
	}
}

func MultiInsertCDR(cdrs []*VoipRestoredCdr) {
	if len(cdrs) == 0 {
		return
	}

	log.Debugf("%d CDRs inserted", len(cdrs))

	sql := "INSERT INTO voip_restored_cdr (call_id,caller_ip,caller_port,callee_ip,callee_port,caller_num,callee_num,caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration) VALUES "
	buf := bytes.Buffer{}
	buf.Write([]byte(sql))
	for _, cdr := range cdrs {
		buf.WriteString(fmt.Sprintf(`("%s","%s",%d,"%s",%d,"%s","%s","%s","%s","%s","%s","%s","%s",%d),`,
			cdr.CallId, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
			cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration))
	}
	// 替换最后一个','
	buf.Bytes()[buf.Len()-1] = ';'

	_, err := orm.NewOrm().Raw(buf.String()).Exec()
	if err != nil {
		log.Error(err)
	}
}

func LogCDR(cdr *VoipRestoredCdr) {
	asyncDao.LogCDR(cdr)
}
