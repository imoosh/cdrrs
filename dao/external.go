package dao

import (
	"bytes"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"centnet-cdrrs/prot/udp"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
)

var errNotFound = errors.New("phone position not found")

type UnpackedMessage struct {
	EventId   string
	EventTime string
	UDP       *udp.UdpMsg
	SIP       *sip.SipMsg
}

func CachePhoneNumberAttribution() error {
	o := orm.NewOrm()
	pp := new(PhonePosition)
	var pps []PhonePosition

	// 查询并缓存手机号码归属地列表
	n, err := o.QueryTable(pp).Filter("phone__isnull", false).GroupBy("phone").All(&pps, "Phone", "Province", "City")
	if err != nil {
		return err
	}
	log.Debugf("Query [%d] results through the 'phone' field in phone_position table", n)
	for _, val := range pps {
		mobilePhoneNumberAttributionMap[val.Phone] = val
	}

	// 查询并缓存座机号码归属地列表
	n, err = o.QueryTable(pp).Filter("code1__isnull", false).GroupBy("code1").All(&pps, "Code1", "Province", "City")
	if err != nil {
		return err
	}
	log.Debugf("Query [%d] results through the 'code1' field in phone_position table", n)
	for _, val := range pps {
		fixedPhoneNumberAttributionMap[val.Code1] = val
		log.Debug(val)
	}

	return nil
}

// 固话号码获取归属地
// n: 3、4：通过前3位或前4位获取归属地，为-1时，前3位或4位都尝试获取
func GetPositionByFixedPhoneNumber(num string, n int) (PhonePosition, error) {
	if len(num) != 11 && len(num) != 12 {
		//log.Error("invalid num number:", num)
		return PhonePosition{}, errNotFound
	}

	if n == 3 || n == -1 {
		if pp, ok := fixedPhoneNumberAttributionMap[num[0:3]]; ok {
			return pp.(PhonePosition), nil
		}
	}

	if n == 4 || n == -1 {
		if pp, ok := fixedPhoneNumberAttributionMap[num[0:4]]; ok {
			return pp.(PhonePosition), nil
		}
	}

	return PhonePosition{}, errNotFound
}

// 手机号码获取归属地
func GetPositionByMobilePhoneNumber(num string) (PhonePosition, error) {
	if len(num) != 11 {
		//log.Error("invalid num number:", num)
		return PhonePosition{}, errNotFound
	}

	if pp, ok := mobilePhoneNumberAttributionMap[num[0:7]]; ok {
		return pp.(PhonePosition), nil
	}
	return PhonePosition{}, errNotFound
}

func InsertCDR(cdr *VoipRestoredCdr) {
	sql := fmt.Sprintf(`
insert into voip_restored_cdr (call_id,uuid,caller_ip,caller_port,callee_ip,callee_port,caller_num,
callee_num,caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration)
        values("%s","%s","%s",%d,"%s",%d,"%s","%s","%s","%s","%s","%s",%d,%d,%d)`,
		cdr.CallId, cdr.Uuid, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
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

	sql := "INSERT INTO voip_restored_cdr (call_id,uuid,caller_ip,caller_port,callee_ip,callee_port,caller_num,callee_num,caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration, create_time) VALUES "
	buf := bytes.Buffer{}
	buf.Write([]byte(sql))
	for _, cdr := range cdrs {
		buf.WriteString(fmt.Sprintf(`("%s","%s","%s",%d,"%s",%d,"%s","%s","%s","%s","%s","%s",%d,%d,%d,"%s"),`,
			cdr.CallId, cdr.Uuid, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
			cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration, cdr.CreateTime))
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
