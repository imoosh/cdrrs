package dao

import (
	"centnet-cdrrs/common/cache/redis"
	"centnet-cdrrs/common/container/pool"
	xtime "centnet-cdrrs/common/time"
	"centnet-cdrrs/model"
	uuid "github.com/satori/go.uuid"
	"testing"
	"time"
)

var d = &Dao{
	c:     nil,
	db:    nil,
	redis: redis.NewPool(getConfig()),
}

func getConfig() *redis.Config {
	return &redis.Config{
		Config: &pool.Config{
			Active:      10,
			Idle:        2,
			IdleTimeout: xtime.Duration(time.Second * 3),
			WaitTimeout: xtime.Duration(time.Second * 3),
			Wait:        false,
		},
		Name:         "redis",
		Proto:        "tcp",
		Addr:         "127.0.0.1:6379",
		Auth:         "",
		DialTimeout:  xtime.Duration(time.Second * 3),
		ReadTimeout:  xtime.Duration(time.Second * 3),
		WriteTimeout: xtime.Duration(time.Second * 3),
	}
}

func getSipItems(n int) map[string]*model.SipCache {
	var items = make(map[string]*model.SipCache)
	for i := 0; i < n; i++ {
		item := getSipItem()
		items[item.CallId] = item
	}

	return items
}

func getSipItem() *model.SipCache {
	return &model.SipCache{
		CallId:         uuid.NewV4().String(),
		Caller:         "1101000",
		Callee:         "18113007500",
		SrcIP:          "192.168.1.25",
		DestIP:         "192.168.1.32",
		SrcPort:        8080,
		DestPort:       5060,
		CallerDevice:   "sip client",
		CalleeDevice:   "sip server",
		ConnectTime:    "20201123120800",
		DisconnectTime: "20201123120852",
	}
}

func TestDao_SGetSipItem(t *testing.T) {
	data := getSipItem()
	d.CacheSipItem(data)

	reply, err := d.GetSipItem(data.CallId)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(reply)
}

func TestDao_SGetSipItems(t *testing.T) {
	data := getSipItems(10000)
	d.CacheSipItems(data)

	var k []string
	for _, v := range data {
		k = append(k, v.CallId)
	}
	reply, err := d.GetSipItems(k)
	if err != nil {
		t.Log(err)
		return
	}
	for _, v := range reply {
		t.Log(v)
	}
}

func TestExpiredSipItemKeys(t *testing.T) {
	data := getSipItems(10000)
	d.CacheSipItems(data)

	var ids []string
	for _, v := range data {
		ids = append(ids, v.CallId)
	}

	key := time.Now().Format("20060102150405")
	d.CacheExpiredSipItemKeysSet(key, ids)

	aa, _ := d.GetExpiredSipItemKeysSetMembers(key)
	t.Log(aa)

	d.DelSipItems(aa)
	d.DelExpiredSipItemKeysSets([]string{key})
}
