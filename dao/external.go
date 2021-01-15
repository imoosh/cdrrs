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

var (
	cdrsCount   uint64 = 0
	sqlBuffer          = bytes.NewBuffer(make([]byte, 10<<20))
	sqlWords           = "(id,caller_ip,caller_port,callee_ip,callee_port,caller_num,callee_num,caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration, create_time) VALUES "
	errNotFound        = errors.New("phone position not found")
)

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
		fixedNumberPositionMap[val.Code1] = val
	}
	log.Debugf("fixedNumberPositionMap cached %d items (%d queried)", len(fixedNumberPositionMap), n)

	// 查询并缓存手机号码归属地列表
	n, err = o.QueryTable(pp).Filter("phone__isnull", false).All(&pps, "Phone", "Province", "City")
	if err != nil {
		return err
	}
	for _, val := range pps {
		mobileNumberPositionMap[val.Phone] = val
	}
	log.Debugf("mobileNumberPositionMap cached %d items (%d queried)", len(mobileNumberPositionMap), n)

	return nil
}

// 固话号码获取归属地
func QueryFixedNumberPosition(num string) (PhonePosition, error) {
	length := len(num)
	if length != 11 && length != 12 {
		return PhonePosition{}, errNotFound
	}

	// 若长度为11，先尝试识别前三位归属，再尝试前四位归属
	if length == 11 {
		if pp, ok := fixedNumberPositionMap[num[0:3]]; ok {
			return pp.(PhonePosition), nil
		}
	}

	// 若长度为12，尝试获取前四位归属
	if pp, ok := fixedNumberPositionMap[num[0:4]]; ok {
		return pp.(PhonePosition), nil
	}

	return PhonePosition{}, errNotFound
}

func ValidateFixedNumber(num string) bool {
	length := len(num)
	if length != 11 && length != 12 {
		return false
	}

	// 若长度为11，先尝试识别前三位归属，再尝试前四位归属
	if length == 11 {
		if _, ok := fixedNumberPositionMap[num[0:3]]; ok {
			return true
		}
	}

	// 若长度为12，尝试获取前四位归属
	if _, ok := fixedNumberPositionMap[num[0:4]]; ok {
		return true
	}

	return false
}

// 手机号码获取归属地
func QueryMobileNumberPosition(num string) (PhonePosition, error) {
	if len(num) != 11 {
		//log.Error("invalid num number:", num)
		return PhonePosition{}, errNotFound
	}

	if pp, ok := mobileNumberPositionMap[num[0:7]]; ok {
		return pp.(PhonePosition), nil
	}
	return PhonePosition{}, errNotFound
}

func ValidateMobileNumber(num string) bool {
	if len(num) != 11 {
		return false
	}

	if _, ok := mobileNumberPositionMap[num[0:7]]; ok {
		return true
	}
	return false
}

func CreateTable(tableName string) {
	sql := "CREATE TABLE IF NOT EXISTS `" + tableName + "`" + " (" +
		"`id` bigint(20) NOT NULL COMMENT '话单唯一ID'," +
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
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;"

	_, err := orm.NewOrm().Raw(sql).Exec()
	if err != nil {
		log.Error(err)
	}
}

func MultiInsertCDR(tableName string, cdrs []*VoipCDR) {
	if len(cdrs) == 0 {
		return
	}

	sqlBuffer.Reset()
	sqlBuffer.WriteString("INSERT INTO " + tableName + sqlWords)
	for _, cdr := range cdrs {
		sqlBuffer.WriteString(fmt.Sprintf(`(%d,'%s',%d,'%s',%d,'%s','%s','%s','%s','%s','%s',%d,%d,%d,'%s'),`,
			cdr.Id, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
			cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration, cdr.CreateTime))
		// 用完回收
		cdr.Free()
	}
	// 替换最后一个','为';'
	sqlBuffer.Bytes()[sqlBuffer.Len()-1] = ';'

	_, err := orm.NewOrm().Raw(sqlBuffer.String()).Exec()
	if err != nil {
		log.Error(err)
	}

	cdrsCount = cdrsCount + uint64(len(cdrs))
	log.Debugf("%4d CDRs (total %d) -> '%s'", len(cdrs), cdrsCount, tableName)
}

func InsertCDR(cdr *VoipCDR) {
	asyncDao.LogCDR(cdr)
}
