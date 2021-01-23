package service

import (
    "centnet-cdrrs/common/cache/redis"
    "context"
    "fmt"
    "sync"
)

type client struct {
    c    *redis.Config
    pool *redis.Pool

    mu sync.Mutex
}

func (cli *client) HandleExpiredRedisCache() {
    c, err := redis.Dial(cli.c.Proto, cli.c.Addr, redis.DialReadTimeout(0))
    if err != nil {
        fmt.Println(err)
    }
    conn := redis.PubSubConn{Conn: c}
    defer conn.Close()

    err = conn.Subscribe("__keyevent@0__:expired")
    if err != nil {
        fmt.Println(err)
    }

    for {
        switch n := conn.Receive().(type) {
        case redis.Message:
            fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
        case redis.PMessage:
            fmt.Printf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
        case redis.Subscription:
            fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
            if n.Count == 0 {
                return
            }
        case error:
            fmt.Println(err)
            return
        }
    }
}

func NewClient(c *redis.Config) *client {
    return &client{
        c:    c,
        pool: redis.NewPool(c),
        mu:   sync.Mutex{},
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

func (cli *client) querySemiFinishedCDR(k string) (interface{}, error ){
    c := cli.pool.Get(context.TODO())
    defer c.Close()

    return c.Do("GET", k)
}
