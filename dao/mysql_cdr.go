package dao

import (
    "bytes"
    "centnet-cdrrs/common/log"
    "errors"
    "fmt"
)

var (
    cdrsCount   uint64 = 0
    sqlBuffer          = bytes.NewBuffer(make([]byte, 10<<20))
    sqlWords           = "(id,caller_ip,caller_port,callee_ip,callee_port,caller_num,callee_num,caller_device,callee_device,callee_province,callee_city,connect_time,disconnect_time,duration, create_time) VALUES "
    errNotFound        = errors.New("phone position not found")
)

type VoipCDR struct {
    Id             int64  `json:"id"`
    CallerIp       string `json:"callerIp"`
    CallerPort     int    `json:"callerPort"`
    CalleeIp       string `json:"calleeIp"`
    CalleePort     int    `json:"calleePort"`
    CallerNum      string `json:"callerNum"`
    CalleeNum      string `json:"calleeNum"`
    CallerDevice   string `json:"callerDevice"`
    CalleeDevice   string `json:"calleeDevice"`
    CalleeProvince string `json:"calleeProvince"`
    CalleeCity     string `json:"calleeCity"`
    ConnectTime    int64  `json:"connectTime"`
    DisconnectTime int64  `json:"disconnectTime"`
    Duration       int    `json:"duration"`
    FraudType      string `json:"fraudType"`
    CreateTime     string `json:"createTime"`
    TableName      string `json:"tableName" orm:"-"`
}

func (d *Dao) CreateTable(tableName string) {
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
        "PRIMARY KEY (`id`)," +
        "UNIQUE KEY `cdr_index` (`id`) USING BTREE" +
        ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;"

    //d.db.Migrator().CreateTable()
    db := d.db.Model(&VoipCDR{})
    err := db.Exec(sql).Error
    if err != nil {
        log.Error(err)
    }
}

func (d *Dao) InsertCDR(tableName string, cdr *VoipCDR) {
    sql := "INSERT INTO " + tableName + sqlWords
    sql += fmt.Sprintf(`(%d,'%s',%d,'%s',%d,'%s','%s','%s','%s','%s','%s',%d,%d,%d,'%s'),`,
        cdr.Id, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
        cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration, cdr.CreateTime)

    db := d.db.Model(&VoipCDR{})
    err := db.Exec(sqlBuffer.String()).Error
    if err != nil {
        log.Error(err)
    }
}

func (d *Dao) MultiInsertCDR(tableName string, cdrs []*VoipCDR) {
    if len(tableName) == 0 || len(cdrs) == 0 {
        return
    }

    sqlBuffer.Reset()
    sqlBuffer.WriteString("INSERT INTO " + tableName + sqlWords)
    for _, cdr := range cdrs {
        sqlBuffer.WriteString(fmt.Sprintf(`(%d,'%s',%d,'%s',%d,'%s','%s','%s','%s','%s','%s',%d,%d,%d,'%s'),`,
            cdr.Id, cdr.CallerIp, cdr.CallerPort, cdr.CalleeIp, cdr.CalleePort, cdr.CallerNum, cdr.CalleeNum, cdr.CallerDevice,
            cdr.CalleeDevice, cdr.CalleeProvince, cdr.CalleeCity, cdr.ConnectTime, cdr.DisconnectTime, cdr.Duration, cdr.CreateTime))
    }
    // 替换最后一个','为';'
    sqlBuffer.Bytes()[sqlBuffer.Len()-1] = ';'

    db := d.db.Model(&VoipCDR{})
    err := db.Exec(sqlBuffer.String()).Error
    if err != nil {
        log.Error(err)
    }

    cdrsCount = cdrsCount + uint64(len(cdrs))
    log.Debugf("%5d CDRs (total %d) -> '%s'", len(cdrs), cdrsCount, tableName)
}
