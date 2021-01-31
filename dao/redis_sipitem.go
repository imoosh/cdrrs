package dao

import (
    "centnet-cdrrs/common/cache/redis"
    "centnet-cdrrs/common/log"
    "centnet-cdrrs/model"
    "context"
    "encoding/json"
)

func (d *Dao) CacheSipItem(data *model.SipItem) {
    var (
        conn = d.redis.Get(context.TODO())
        err  error
        bs   []byte
    )
    defer conn.Close()

    if bs, err = json.Marshal(data); err != nil {
        log.Errorf("json.Marshal error(%v)", err)
        return
    }

    _, err = conn.Do("SET", data.CallId, bs)
    if err != nil {
        log.Error(err)
        return
    }
}

func (d *Dao) CacheSipItems(data map[string]*model.SipItem) {
    var (
        conn = d.redis.Get(context.TODO())
        args = redis.Args{}
        err  error
        bs   []byte
    )
    defer conn.Close()

    for k, v := range data {
        if bs, err = json.Marshal(v); err != nil {
            log.Errorf("json.Marshal err(%v)", err)
            continue
        }
        args = args.Add(k).Add(string(bs))
    }
    if err = conn.Send("MSET", args...); err != nil {
        log.Errorf("conn.Send(MSET) error(%v)", err)
        return
    }
    if err = conn.Flush(); err != nil {
        log.Errorf("conn.Flush() error(%v)", err)
        return
    }
}

func (d *Dao) GetSipItem(k string) (item *model.SipItem, err error) {
    var (
        bs   []byte
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    if bs, err = redis.Bytes(conn.Do("GET", k)); err != nil {
        if err == redis.ErrNil {
            return nil, nil
        }
        log.Error(err)
        return nil, err
    }
    item = model.NewSipItem()
    if err = json.Unmarshal(bs, item); err != nil {
        item.Free()
        log.Errorf("json.Unmarshal error(%v)", err)
        return nil, err
    }
    item.CallId = k

    return item, nil
}

func (d *Dao) GetSipItems(ids []string) (res map[string]*model.SipItem, err error) {
    var (
        bss  [][]byte
        args = redis.Args{}
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    for _, v := range ids {
        args = args.Add(v)
    }
    if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
        if err == redis.ErrNil {
            return nil, nil
        } else {
            log.Errorf("conn.Do(MGET) error(%v)", err)
        }
        return
    }

    res = make(map[string]*model.SipItem, len(bss))
    for i, bs := range bss {
        if bs == nil {
            res[ids[i]] = nil
            continue
        }

        item := model.NewSipItem()
        if err = json.Unmarshal(bs, item); err != nil {
            log.Errorf("json.Unmarshal error(%v)", err)

            item.Free()
            res[ids[i]] = nil
            err = nil

            continue
        }

        item.CallId = ids[i]
        res[ids[i]] = item
    }

    return
}

func (d *Dao) DelSipItems(ids []string) {
    var (
        args = redis.Args{}
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()
    for _, v := range ids {
        args = args.Add(v)
    }

    _, err := conn.Do("DEL", args...)
    if err != nil {
        log.Error(err)
    }
}

func (d *Dao) CacheExpiredSipItemKeysSet(key string, ids []string) {
    var (
        err  error
        args = redis.Args{}
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    args = args.Add(key)
    for _, v := range ids {
        args = args.Add(v)
    }

    if nil != selectDB(conn, 1) {
        return
    }

    _, err = conn.Do("SADD", args...)
    if err != nil {
        log.Errorf("conn.Send(SADD) error(%v)", err)
        return
    }

    if nil != selectDB(conn, 0) {
        return
    }
}

func (d *Dao) GetExpiredSipItemKeysSets() (sids []string) {
    var (
        startCursor int
        endCursor   int
        conn        = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    if nil != selectDB(conn, 1) {
        return
    }

    endCursor = 1
    for endCursor != 0 {
        var (
            err         error
            trunkedSids []string
            values      []interface{}
        )
        if values, err = redis.Values(conn.Do("SCAN", startCursor)); err != nil {
            log.Errorf("conn.Do(SCAN) error(%v)", err)
            return
        }
        if _, err = redis.Scan(values, &endCursor, &trunkedSids); err != nil {
            log.Errorf("redis.Scan(%v) error(%v)", values, err)
            return
        }
        startCursor = endCursor
        sids = append(sids, trunkedSids...)
    }

    if nil != selectDB(conn, 0) {
        return
    }
    return
}

func (d *Dao) DelExpiredSipItemKeysSets(sids []string) {
    var (
        err  error
        args = redis.Args{}
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    for _, v := range sids {
        args = args.Add(v)
    }

    if nil != selectDB(conn, 1) {
        return
    }

    _, err = conn.Do("DEL", args...)
    if err != nil {
        log.Errorf("conn.Send(DEL) error(%v)", err)
        return
    }

    if nil != selectDB(conn, 0) {
        return
    }
}

func (d *Dao) GetExpiredSipItemKeysSetMembers(k string) (mem []string, err error) {
    var (
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    if nil != selectDB(conn, 1) {
        return
    }

    if mem, err = redis.Strings(conn.Do("SMEMBERS", k)); err != nil {
        if err == redis.ErrNil {
            return nil, nil
        }
        log.Errorf("conn.Do(SMEMBERS) error(%v)", err)
        return nil, err
    }

    if nil != selectDB(conn, 0) {
        return
    }

    return
}

func (d *Dao) GetExpiredSipItemKeysSetSize(key string) (n int) {
    var (
        err  error
        conn = d.redis.Get(context.TODO())
    )
    defer conn.Close()

    if nil != selectDB(conn, 1) {
        return
    }

    if n, err = redis.Int(conn.Do("SCARD", key)); err != nil {
        log.Error("conn.Do(SCARD) error(%v)", err)
        return 0
    }

    if nil != selectDB(conn, 0) {
        return
    }

    return n
}

func selectDB(conn redis.Conn, index int) (err error) {
    _, err = conn.Do("SELECT", index)
    if err != nil {
        log.Errorf("conn.Send(SELECT) error(%v)", err)
    }
    return
}
