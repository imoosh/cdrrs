package dao

import (
	"centnet-cdrrs/common/database/orm"
	"centnet-cdrrs/conf"
	"testing"
)

func TestCachePhonePosition(t *testing.T) {
	dao := New(&conf.Config{
		Logging: nil,
		Kafka:   nil,
		ORM: &orm.Config{
			DSN:         "centnet_cdrrs:123456@tcp(192.168.1.205:3306)/centnet_cdrrs?timeout=10s&readTimeout=10s&writeTimeout=10s&parseTime=true&loc=Local&charset=utf8,utf8mb4",
			Active:      10,
			Idle:        100,
			IdleTimeout: 600,
		},
	})

	dao.CacheMobilePosition()
	dao.CacheFixedPosition()
}
