package conmap

import (
	cmap "centnet-cdrrs/adapter/cache/cmap"
	"github.com/RussellLuo/timingwheel"
	"time"
)

type ExpireMap struct {
	cm *cmap.ConcurrentMap
	tw *timingwheel.TimingWheel
}

func NewExpireMap() *ExpireMap {

	expireMap := &ExpireMap{
		cm: cmap.CreateConcurrentMap(1000),
		tw: timingwheel.NewTimingWheel(time.Second, 60),
	}
	expireMap.tw.Start()

	return expireMap
}

func (c *ExpireMap) Get(key interface{}) (interface{}, bool) {
	return c.cm.Get(cmap.StrKey(key.(string)))
}

func (c *ExpireMap) Set(key interface{}, value interface{}) (interface{}, bool) {
	return c.cm.Set(cmap.StrKey(key.(string)), value)
}

func (c *ExpireMap) SetWithExpire(key interface{}, value interface{}, expire time.Duration) (interface{}, bool) {
	v, ok := c.Set(key, value)
	if ok {
		c.tw.AfterFunc(expire, func() { c.Del(key) })
	}
	return v, ok
}

func (c *ExpireMap) Del(key interface{}) {
	c.cm.Del(cmap.StrKey(key.(string)))
}
