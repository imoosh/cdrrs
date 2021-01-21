package service

import (
	"centnet-cdrrs/common/cache/redis"
	"centnet-cdrrs/common/log"
	"context"
)

func RedisTest() {
	pool := redis.NewPool(getConfig())
	_, _ = pool.Get(context.TODO()).Do("SET", "abc", "123")
}

func handleExpiredRedisCache() {
	pool := redis.NewPool(getConfig())

	conn := redis.PubSubConn{Conn: pool.Get(context.TODO())}
	defer conn.Close()

	err := conn.Subscribe("__keyevent@0__:expired")
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
