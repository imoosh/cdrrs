package cache

import (
    cmap "centnet-cdrrs/adapter/cache/cmap"
    "time"
)

type ExpireMap struct {
    cm *cmap.ConcurrentMap
}

func NewExpireMap() *ExpireMap {
    duration := 60 * time.Minute
    return &ExpireMap{
        cm: cmap.CreateConcurrentMap(100, duration),
    }
}

func (c *ExpireMap) Get(key interface{}) (interface{}, bool) {
    return c.cm.Get(cmap.StrKey(key.(string)))
}

func (c *ExpireMap) SetWithExpire(key interface{}, value interface{}) (interface{}, bool) {
    return c.cm.SetEx(cmap.StrKey(key.(string)), value)
}

func (c *ExpireMap) Del(key interface{}) {
    c.cm.Del(cmap.StrKey(key.(string)))
}
