package redis

import (
	"centnet-cdrrs/library/log"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

var redisPool *redis.Pool
var Conf *Config

func newPool(conf *Config) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: time.Duration(conf.IdleTimeout) * time.Second,
		Wait:        true, // 超过最大连接，等待或报错
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conf.Host)
			if err != nil {
				return nil, err
			}
			if len(conf.Password) != 0 {
				if _, err := c.Do("AUTH", conf.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func Put(k, v string) error {
	conn := redisPool.Get()
	defer conn.Close()
	if _, err := conn.Do("SET", k, v); err != nil {
		log.Error(err)
		return errors.New(fmt.Sprintf("redis command: SET %s %s error", k, v))
	}
	return nil
}

func PutWithExpire(k, v string, expire int) error {
	conn := redisPool.Get()
	defer conn.Close()
	if _, err := conn.Do("SETEX", k, expire, v); err != nil {
		log.Error(err)
		return errors.New(fmt.Sprintf("redis command: SETEX %s %d %s error", k, expire, v))
	}
	return nil
}

func Delete(k string) error {
	conn := redisPool.Get()
	defer conn.Close()
	if _, err := conn.Do("DEL", k); err != nil {
		log.Error(err)
		return errors.New(fmt.Sprintf("redis command: DEL %s error", k))
	}
	return nil
}

func Get(k string) (string, error) {
	conn := redisPool.Get()
	defer conn.Close()
	v, err := redis.String(conn.Do("GET", k))
	if err == redis.ErrNil {
		return "", nil
	}
	if err != nil {
		log.Error(err)
		return "", errors.New(fmt.Sprintf("redis command: GET %s error", k))
	}

	return v, nil
}

func InitRedisPool(c *Config) error {
	redisPool = newPool(c)
	Conf = c
	return nil
}
