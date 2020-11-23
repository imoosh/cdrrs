package redis

import (
    "centnet-cdrrs/library/log"
    "errors"
    "fmt"
    "github.com/gomodule/redigo/redis"
    "time"
)

var RedisConn *Conn
var redisClient *redis.Pool

type Config struct {
    Host        string
    CacheExpire int
}

type Conn struct {
    conn redis.Conn
    Conf Config
}

func NewRedisConn(c *Config) (*Conn, error) {
    conn, err := redis.Dial("tcp", c.Host)
    if err != nil {
        log.Error(err)
        return nil, errors.New(fmt.Sprintf("redis.Dial %s failed", c.Host))
    }
    return &Conn{conn: conn, Conf: *c}, nil
}

func (rc *Conn) Close() {
    rc.conn.Close()
}

func (rc *Conn) Put(k, v string) error {
    if _, err := rc.conn.Do("SET", k, v); err != nil {
        log.Error(err)
        return errors.New(fmt.Sprintf("redis command: SET %s %s error", k, v))
    }
    return nil
}

func (rc *Conn) PutWithExpire(k, v string, expire int) error {
    if _, err := rc.conn.Do("SETEX", k, expire, v); err != nil {
        log.Error(err)
        return errors.New(fmt.Sprintf("redis command: SETEX %s %d %s error", k, expire, v))
    }
    return nil
}

func (rc *Conn) Delete(k string) error {
    if _, err := rc.conn.Do("DEL", k); err != nil {
        log.Error(err)
        return errors.New(fmt.Sprintf("redis command: DEL %s error", k))
    }
    return nil
}

func (rc *Conn) Get(k string) (string, error) {
    v, err := redis.String(rc.conn.Do("GET", k))
    if err == redis.ErrNil {
        return "", nil
    }
    if err != nil {
        log.Error(err)
        return "", errors.New(fmt.Sprintf("redis command: GET %s error", k))
    }

    return v, nil
}

func (rc *Conn) Test() bool {
    err := rc.Put("abc", "123")
    if err != nil {
        log.Errorf("redis test: 'put abc 123' error")
        return false
    }
    log.Debug("redis test: 'put abc 123' ok")

    v, err := rc.Get("abc")
    if err != nil {
        log.Errorf("redis test: 'get abc' error")
        return false
    }
    log.Debugf("redis test: 'get abc %s' ok", v)

    err = rc.Delete("abc")
    if err != nil {
        log.Errorf("redis test: 'del abc' error")
        return false
    }
    log.Debugf("redis test: 'del abc' ok")

    err = rc.PutWithExpire("abc", "123", 1)
    if err != nil {
        log.Errorf("redis test: 'set abc 1 123' error")
        return false
    }
    //time.Sleep(time.Second * 2)
    v, err = rc.Get("abc")
    if err != nil {
        log.Errorf("redis test: 'get abc' error")
    }

    return true
}

func Init(c *Config) error {
    var err error
    RedisConn, err = NewRedisConn(c)
    if err != nil {
        log.Error(err)
        return errors.New("new redis connection error")
    }

    if !RedisConn.Test() {
        return errors.New("redis test failed")
    }

    return nil
}

func initRedisPool(server string, password string) *redis.Pool {
    return &redis.Pool{
        MaxIdle:     2, //空闲数
        IdleTimeout: 240 * time.Second,
        MaxActive:   3, //最大数
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", server)
            if err != nil {
                return nil, err
            }
            if password != "" {
                if _, err := c.Do("AUTH", password); err != nil {
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
