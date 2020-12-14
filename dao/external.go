package dao

import (
	"bytes"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model/prot/sip"
	"centnet-cdrrs/model/prot/udp"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
)

var cdrsCount uint64 = 0
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

	log.Debug("")
	// 查询并缓存座机号码归属地列表
	n, err := o.QueryTable(pp).Filter("code1__isnull", false).GroupBy("code1").All(&pps, "Code1", "Province", "City")
	if err != nil {
		return err
	}
	for _, val := range pps {
		fixedPhoneNumberAttributionMap[val.Code1] = val
	}
	log.Debugf("fixedPhoneNumberAttributionMap cached %d items (%d queried)", len(fixedPhoneNumberAttributionMap), n)

	// 查询并缓存手机号码归属地列表
	n, err = o.QueryTable(pp).Filter("phone__isnull", false).All(&pps, "Phone", "Province", "City")
	if err != nil {
		return err
	}
	for _, val := range pps {
		mobilePhoneNumberAttributionMap[val.Phone] = val
	}
	log.Debugf("mobilePhoneNumberAttributionMap cached %d items (%d queried)", len(mobilePhoneNumberAttributionMap), n)

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

func CreateTable(tableName string) {
	sql := "CREATE TABLE IF NOT EXISTS `" + tableName + "`" + " (" +
		"`id` bigint(20) NOT NULL AUTO_INCREMENT," +
		"`uuid` varchar(64) NOT NULL COMMENT '话单唯一ID'," +
		"`call_id` varchar(128) DEFAULT NULL COMMENT '通话ID'," +
		"`caller_ip` varchar(64) DEFAULT NULL COMMENT '主叫IP', " +
		"`caller_port` int(8) DEFAULT NULL COMMENT '主叫端口'," +
		"`callee_ip` varchar(64) DEFAULT NULL COMMENT '被叫IP'," +
		"`callee_port` int(8) DEFAULT NULL COMMENT '被叫端口'," +
		"`caller_num` varchar(64) DEFAULT NULL COMMENT '主叫号码'," +
		"`callee_num` varchar(64) DEFAULT NULL COMMENT '被叫号码'," +
		"`caller_device` varchar(128) DEFAULT NULL COMMENT '主叫设备名'," +
		"`callee_device` varchar(128) DEFAULT NULL COMMENT '被叫设备名'," +
		"`callee_province` varchar(64) DEFAULT NULL COMMENT '被叫所属省'," +
		"`callee_city` varchar(64) DEFAULT NULL COMMENT '被叫所属市'," +
		"`connect_time` int(11) DEFAULT NULL COMMENT '通话开始时间'," +
		"`disconnect_time` int(11) DEFAULT NULL COMMENT '通话结束时间'," +
		"`duration` int(8) DEFAULT '0' COMMENT '通话时长'," +
		"`fraud_type` varchar(32) DEFAULT NULL COMMENT '诈骗类型'," +
		"`create_time` datetime DEFAULT NULL COMMENT '生成时间'," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY `cdr_uuid` (`uuid`) USING BTREE" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;"

	_, err := orm.NewOrm().Raw(sql).Exec()
	if err != nil {
		log.Error(err)
	}
}

func MultiInsertCDR(tableName string, cdrs []*VoipRestoredCdr) {
	if len(cdrs) == 0 {
		return
	}

	sql := "INSERT INTO " + tableName + "(call_id,uuid,caller_ip,caller_port,callee_ip,callee_port,caller_num,callee_num," +
		"caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration, create_time) VALUES "

	buf := bytes.Buffer{}
	buf.Write([]byte(sql))
	for _, cdr := range cdrs {
		buf.WriteString(fmt.Sprintf(`("%s","%s","%s",%d,"%s",%d,"%s","%s","%s","%s","%s","%s",%d,%d,%d,"%s"),`,
			cdr.CallId, cdr.Uuid, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
			cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration, cdr.CreateTime))
	}
	// 替换最后一个','为';'
	buf.Bytes()[buf.Len()-1] = ';'

	_, err := orm.NewOrm().Raw(buf.String()).Exec()
	if err != nil {
		log.Error(err)
	}

	cdrsCount = cdrsCount + uint64(len(cdrs))
	log.Debugf("%4d CDRs (total %d) -> '%s'", len(cdrs), cdrsCount, tableName)
}

func LogCDR(cdr *VoipRestoredCdr) {
	asyncDao.LogCDR(cdr)
}
