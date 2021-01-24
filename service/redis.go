package service

import (
	"centnet-cdrrs/common/cache/redis"
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/conf"
	"context"
	"sync"
)

type client struct {
	c    *redis.Config
	pool *redis.Pool

	mu sync.Mutex
}

func handleExpiredRedisCache(conf *conf.Config) {
	c, err := redis.Dial(conf.Redis.Proto, conf.Redis.Addr, redis.DialReadTimeout(0))
	if err != nil {
		log.Error(err)
	}
	conn := redis.PubSubConn{Conn: c}
	defer conn.Close()

	err = conn.Subscribe("__keyevent@0__:expired")
	if err != nil {
		log.Error(err)
	}

	for {
		switch n := conn.Receive().(type) {
		case redis.Message:
			log.Debugf("Message: %s %s\n", n.Channel, n.Data)
		case redis.PMessage:
			log.Debugf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
		case redis.Subscription:
			log.Debugf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
			if n.Count == 0 {
				return
			}
		case error:
			log.Debugf("error: %v\n", n)
			return
		}
	}
}

func (cli *client) pushSemiFinishedCDRToRedis(args ...interface{}) {
	c := cli.pool.Get(context.TODO())
	defer c.Close()

	c.Send("SET", args...)
	c.Flush()
}

func (cli *client) pushSemiFinishedCDRToRedisWithExpired(args ...interface{}) {
	c := cli.pool.Get(context.TODO())
	defer c.Close()

	c.Send("SETEX", args...)
	c.Flush()
}

func (cli *client) querySemiFinishedCDR(key string) {
	c := cli.pool.Get(context.TODO())
	defer c.Close()

	reply, err := c.Do("SCAN", 0, "MATCH", key+"*", "COUNT", 1)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug(reply)
}

func (cli *client) deleteSemiFinishedCDR(key string) {
	c := cli.pool.Get(context.TODO())
	defer c.Close()

	c.Do("DEL")
}
