package redis

import (
	"centnet-cdrrs/library/log"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

var RedisConn *Conn

type Config struct {
	Host        string
	CacheExpire int
}

type Conn struct {
	conn redis.Conn
}

func NewRedisConn(c *Config) (*Conn, error) {
	conn, err := redis.Dial("tcp", c.Host)
	if err != nil {
		log.Error(err)
		return nil, errors.New(fmt.Sprintf("redis.Dial %s failed", c.Host))
	}
	return &Conn{conn: conn}, nil
}

func (rc *Conn) Close() {
	rc.conn.Close()
}

func (rc *Conn) Put(k, v string) {
	if _, err := rc.conn.Do("SET", k, v); err != nil {
		fmt.Println(err)
		//log.Error(err)
	}
}

func (rc *Conn) PutWithExpire(k, v string, expire int) {
	if _, err := rc.conn.Do("SETEX", k, expire, v); err != nil {
		log.Error(err)
	}
}

func (rc *Conn) Delete(k string) {
	if _, err := rc.conn.Do("DEL", k); err != nil {
		log.Error(err)
	}
}

func (rc *Conn) Get(k string) (string, error) {
	v, err := redis.String(rc.conn.Do("GET", k))
	if err != nil {
		log.Error(err)
		return "", errors.New(fmt.Sprintf("command redis:  GET %s error", k))
	}

	return v, nil
}

func Init(c *Config) error {
	var err error
	RedisConn, err = NewRedisConn(c)
	if err != nil {
		log.Error(err)
		return errors.New("new redis connection error")
	}
	return nil
}
