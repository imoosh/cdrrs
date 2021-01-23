package service

import (
    "centnet-cdrrs/common/cache/redis"
    "centnet-cdrrs/common/container/pool"
    xtime "centnet-cdrrs/common/time"
    "testing"
    "time"
)

func getConfig() *redis.Config {
    return &redis.Config{
        Name:  "article",
        Proto: "tcp",
        Addr:  "192.168.1.205:6379",
        Config: &pool.Config{
            Idle:   2,
            Active: 5,
        },
        DialTimeout:  xtime.Duration(int64(1000000000)),
        ReadTimeout:  xtime.Duration(int64(1000000000)),
        WriteTimeout: xtime.Duration(int64(1000000000)),
    }
}

func TestHandleExpiredRedisCache(t *testing.T) {

    cli := NewClient(getConfig())
    go cli.HandleExpiredRedisCache()

    cli.pushSemiFinishedCDRToRedisWithExpired("abc", 5, "123")
    time.Sleep(time.Second)
    t.Log(cli.querySemiFinishedCDR("abc"))
    time.Sleep(time.Second*5)
    t.Log(cli.querySemiFinishedCDR("abc"))

    select {}
}
