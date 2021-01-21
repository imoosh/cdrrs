package redis

import (
    "context"
    "sync"
    "testing"
    "time"

    "centnet-cdrrs/common/container/pool"
    xtime "centnet-cdrrs/common/time"
)

var p *Pool
var config *Config

func init() {
    config = getConfig()
    p = NewPool(config)
}

func getConfig() (c *Config) {
    c = &Config{
        Name:         "test",
        Proto:        "tcp",
        Addr:         "192.168.1.205:6379",
        DialTimeout:  xtime.Duration(time.Second),
        ReadTimeout:  xtime.Duration(time.Second*10),
        WriteTimeout: xtime.Duration(time.Second),
    }
    c.Config = &pool.Config{
        Active:      20,
        Idle:        2,
        IdleTimeout: xtime.Duration(90 * time.Second),
    }
    return
}

func TestRedis(t *testing.T) {
    testSet(t, p)
    testSend(t, p)
    testGet(t, p)
    testErr(t, p)
    if err := p.Close(); err != nil {
        t.Errorf("redis: close error(%v)", err)
    }
    conn, err := NewConn(config)
    if err != nil {
        t.Errorf("redis: new conn error(%v)", err)
    }
    if err := conn.Close(); err != nil {
        t.Errorf("redis: close error(%v)", err)
    }
}

func testSet(t *testing.T, p *Pool) {
    var (
        key   = "test"
        value = "test"
        conn  = p.Get(context.TODO())
    )
    defer conn.Close()
    if reply, err := conn.Do("set", key, value); err != nil {
        t.Errorf("redis: conn.Do(SET, %s, %s) error(%v)", key, value, err)
    } else {
        t.Logf("redis: set status: %s", reply)
    }
}

func testSend(t *testing.T, p *Pool) {
    var (
        key    = "test"
        value  = "test"
        expire = 1000
        conn   = p.Get(context.TODO())
    )
    defer conn.Close()
    if err := conn.Send("SET", key, value); err != nil {
        t.Errorf("redis: conn.Send(SET, %s, %s) error(%v)", key, value, err)
    }
    if err := conn.Send("EXPIRE", key, expire); err != nil {
        t.Errorf("redis: conn.Send(EXPIRE key(%s) expire(%d)) error(%v)", key, expire, err)
    }
    if err := conn.Flush(); err != nil {
        t.Errorf("redis: conn.Flush error(%v)", err)
    }
    for i := 0; i < 2; i++ {
        if _, err := conn.Receive(); err != nil {
            t.Errorf("redis: conn.Receive error(%v)", err)
            return
        }
    }
    t.Logf("redis: set value: %s", value)
}

func testGet(t *testing.T, p *Pool) {
    var (
        key  = "test"
        conn = p.Get(context.TODO())
    )
    defer conn.Close()
    if reply, err := conn.Do("GET", key); err != nil {
        t.Errorf("redis: conn.Do(GET, %s) error(%v)", key, err)
    } else {
        t.Logf("redis: get value: %s", reply)
    }
}

func testErr(t *testing.T, p *Pool) {
    conn := p.Get(context.TODO())
    if err := conn.Close(); err != nil {
        t.Errorf("memcache: close error(%v)", err)
    }
    if err := conn.Err(); err == nil {
        t.Errorf("redis: err not nil")
    } else {
        t.Logf("redis: err: %v", err)
    }
}

func BenchmarkMemcache(b *testing.B) {
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            conn := p.Get(context.TODO())
            if err := conn.Close(); err != nil {
                b.Errorf("memcache: close error(%v)", err)
            }
        }
    })
    if err := p.Close(); err != nil {
        b.Errorf("memcache: close error(%v)", err)
    }
}

func TestMSet(t *testing.T) {

    var err error
    conn := p.Get(context.TODO())
    defer conn.Close()

    if err = conn.Send("MSET", "a", "1", "b", "2", "c", "3"); err != nil {
        t.Log(err)
        return
    }
    if err = conn.Flush(); err != nil {
        t.Log(err)
        return
    }
    for i := 0; i < 1; i++ {
        if _, err = conn.Receive(); err != nil {
            t.Logf("conn.Receive() error(%v)", err)
            return
        }
    }
}

func TestMGet(t *testing.T) {

    var err error
    conn := p.Get(context.TODO())
    defer conn.Close()

    if err = conn.Send("MGET", "a", "b", "c"); err != nil {
        t.Log(err)
        return
    }
    if err = conn.Flush(); err != nil {
        t.Log(err)
        return
    }
    for i := 0; i < 1; i++ {
        if v, err := conn.Receive(); err != nil {
            t.Logf("conn.Receive() error(%v)", err)
            return
        } else {
            t.Log(v)
        }
    }
}

func TestPublish(t *testing.T) {
    var wg sync.WaitGroup
    wg.Add(2)

    conn := PubSubConn{p.Get(context.TODO())}
    defer conn.Close()

    conn.Subscribe("__keyevent@0__:expired")

    go func() {
        defer wg.Done()
        for {
            switch n := conn.Receive().(type) {
            case Message:
                t.Logf("Message: %s %s\n", n.Channel, n.Data)
            case PMessage:
                t.Logf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
            case Subscription:
                t.Logf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
                if n.Count == 0 {
                    return
                }
            case error:
                t.Logf("error: %v\n", n)
                return
            default:
                t.Log(n)
            }
        }
    }()

    //publish := func(channel, value interface{}) {
    //    c := p.Get(context.TODO())
    //    defer c.Close()
    //    c.Do("PUBLISH", channel, value)
    //}
    //
    //go func() {
    //    defer wg.Done()
    //    conn.Subscribe("example")
    //    conn.PSubscribe("p*")
    //
    //    publish("example", "hello")
    //    publish("example", "world")
    //    publish("pexample", "foo")
    //    publish("pexample", "bar")
    //
    //    conn.Unsubscribe()
    //    conn.PUnsubscribe()
    //}()

    if _, err := p.Get(context.TODO()).Do("SETEX", "abc", "5", "123"); err != nil {
        t.Log(err)
    }

    wg.Wait()
}
