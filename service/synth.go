package service

import (
    "centnet-cdrrs/common/log"
    "centnet-cdrrs/conf"
    "centnet-cdrrs/dao"
    "centnet-cdrrs/model"
    "centnet-cdrrs/service/cdr"
    "errors"
    "runtime"
    "strings"
    "time"
)

// 话单合成器

var (
    _setIdsTimeFormat   = "20060102150405.000"
    _defaultExpiredFunc = func(interface{}) {}
)

type Synth struct {
    c *conf.Config

    todo    map[string]*model.SipItem
    cdrCh   chan *model.SipItem
    exSynth chan exSynth

    dao    *dao.Dao
    cdrPro *cdr.CDRProducer

    onExpired func(interface{})

    nextPushTime time.Time     // 下次推送缓存时间
    pushPeriod   time.Duration // 由时间间隔决定本地缓存容量
    maxCacheSize int           // 由缓存大小决定本地缓存容量
}

// 超时合成话单
type exSynth struct {
    sid   string
    timer *time.Timer
}

func newSynthesizer(svr *Service) *Synth {
    return &Synth{
        c:            svr.c,
        todo:         make(map[string]*model.SipItem),
        cdrCh:        make(chan *model.SipItem),
        exSynth:      make(chan exSynth, 10000),
        onExpired:    _defaultExpiredFunc,
        pushPeriod:   time.Millisecond * 500,
        maxCacheSize: 10000,
        dao:          svr.dao,
        cdrPro:       svr.cdrPro,
    }
}

func (synth *Synth) Input(item *model.SipItem) {
    synth.cdrCh <- item
}

func (synth *Synth) OnExpired(handle func(interface{})) {
    synth.onExpired = handle
}

func (synth *Synth) handleExpiredKeySet(sid string) {
    var (
        err   error
        keys  []string
        items map[string]*model.SipItem
    )

    // 获取超时集合里所有key
    keys, err = synth.dao.GetExpiredSipItemKeysSetMembers(sid)
    if err != nil || len(keys) == 0 {
        return
    }

    // 获取所有key值
    items, err = synth.dao.GetSipItems(keys)
    if err != nil || len(items) == 0 {
        return
    }

    // TODO: 输出未成功合成的话单
    synth.onExpired(items)

    // 清除redis缓存
    synth.dao.DelSipItems(keys)
    synth.dao.DelExpiredSipItemKeysSets([]string{sid})
}

// 由于软件重启，redis话单缓存失去ids关联，需超时清理掉
func (synth *Synth) unrelatedExpiredKeysSynth() {
    sids := synth.dao.GetExpiredSipItemKeysSets()
    log.Debug("expired sets: ", len(sids))

    var (
        err    error
        t, now time.Time
        d      time.Duration
    )
    for _, sid := range sids {
        if t, err = parseTimeFromSetId(sid); err != nil {
            log.Debugf("parseTimeFromSetId(%s) error(%v)", sid, err)
            continue
        }
        // 还没到超时时间，按实际剩余时间定时，否则立即清理
        now = time.Now()
        if now.Sub(t) < time.Duration(synth.c.CDR.CachedLife) {
            d = time.Duration(synth.c.CDR.CachedLife) - now.Sub(t)
        } else {
            d = time.Second
        }
        synth.exSynth <- exSynth{
            sid:   sid,
            timer: time.NewTimer(d),
        }
    }
}

// 定时处理超时key并输出话单
func (synth *Synth) expiredSynth() {
    n := runtime.NumCPU()
    for i := 0; i < n; i++ {
        go func() {
            for es := range synth.exSynth {
                select {
                case <-es.timer.C:
                    synth.handleExpiredKeySet(es.sid)
                }
            }
        }()
    }
}

//////////////////////////////////////////
func (synth *Synth) sipItemKeys() (ks []string) {
    for k := range synth.todo {
        ks = append(ks, k)
    }
    return
}

func (synth *Synth) sipItemKeysSetId() string {
    return strings.Replace(synth.nextPushTime.Format(_setIdsTimeFormat), ".", "", 1)
}

func parseTimeFromSetId(sid string) (t time.Time, err error) {
    var length = len(sid)
    if length != len(_setIdsTimeFormat)-1 {
        return t, errors.New("invalid time string length")
    }

    stime := sid[:length-3] + "." + sid[length-3:]
    t, err = time.ParseInLocation(_setIdsTimeFormat, stime, time.Local)
    if err != nil {
        log.Errorf("time.Parse error(%v)", err)
    }
    return
}

func (synth *Synth) clearSipItems() {
    for k, v := range synth.todo {
        if v != nil {
            v.Free()
        }
        delete(synth.todo, k)
    }
}

func (synth *Synth) synthesize(a, b *model.SipItem) bool {
    if a.Type == model.SipStatusInvite200OK && b.Type == model.SipStatusBye200OK {
        if a.CallId != b.CallId {
            log.Panicf("a.CallId(%s) != b.CallId(%s)", a.CallId, b.CallId)
        }
        a.DisconnectTime = b.DisconnectTime
        if c := synth.cdrPro.Gen(a.CallId, a); c != nil {
            synth.cdrPro.Put(c)
        }
        return true
    } else if a.Type == model.SipStatusBye200OK && b.Type == model.SipStatusInvite200OK {
        if a.CallId != b.CallId {
            log.Panicf("a.CallId(%s) != b.CallId(%s)", a.CallId, b.CallId)
        }
        b.DisconnectTime = a.DisconnectTime
        if c := synth.cdrPro.Gen(b.CallId, b); c != nil {
            synth.cdrPro.Put(c)
        }
        return true
    }

    return false
}

func (synth *Synth) trySynthesize() {
    keys := synth.sipItemKeys()
    if len(keys) == 0 {
        return
    }

    // 批量查询各报文对应的200OK包
    allQueried, err := synth.dao.GetSipItems(keys)
    if err != nil {
        log.Errorf("synth.dao.GetSipItems error(%v)", err)
        return
    }

    // 依次处理查询结果
    var queriedKeys []string
    for k, v := range synth.todo {
        // 未查询到集中缓存
        if allQueried[k] == nil {
            continue
        }

        // 查询到立即合成话单
        if synth.synthesize(v, allQueried[k]) {
            queriedKeys = append(queriedKeys, v.CallId)
        }
        delete(synth.todo, k)
        v.Free()
        allQueried[k].Free()
    }

    // 删除查询到key（因为已生成话单）
    if len(queriedKeys) != 0 {
        synth.dao.DelSipItems(queriedKeys)
    }
}

func (synth *Synth) append(item *model.SipItem) {
    // 先尝试再本地缓存中匹对数据
    if val, ok := synth.todo[item.CallId]; ok {
        if val.CallId != item.CallId {
            log.Fatalf("val.CallId(%s) != item.CallId(%s)", val.CallId, item.CallId)
        }

        if synth.synthesize(val, item) {
            delete(synth.todo, val.CallId)
            val.Free()
        }
        item.Free()
        return
    }

    synth.todo[item.CallId] = item
}

// 合成话单
func (synth *Synth) realTimeSynth() {
    pushRedis := func() {
        if len(synth.todo) == 0 {
            return
        }
        // 尝试实时合成
        synth.trySynthesize()

        if len(synth.todo) == 0 {
            return
        }
        // 缓存key - value
        synth.dao.CacheSipItems(synth.todo)

        // 缓存key集合
        sid := synth.sipItemKeysSetId()
        ids := synth.sipItemKeys()
        synth.dao.CacheExpiredSipItemKeysSet(sid, ids)

        // 启动超时处理keys机制
        synth.exSynth <- exSynth{
            sid:   sid,
            timer: time.NewTimer(time.Duration(synth.c.CDR.CachedLife)),
        }

        synth.clearSipItems()
    }

    var ctime time.Time // 上次批量处理数据时间
    ticker := time.NewTicker(synth.pushPeriod / 2)
    for {
        select {
        case item := <-synth.cdrCh:
            if len(item.CallId) == 0 {
                item.Free()
                continue
            }

            if synth.needPushRedis() {
                pushRedis()
                ctime = time.Now()
                continue
            }

            synth.append(item)

        case <-ticker.C:
            // 第一个case在pushPeriod周期内未滚动，激活刷新数据流程
            if !ctime.IsZero() && time.Now().Sub(ctime) >= synth.pushPeriod {
                pushRedis()
                ctime = time.Now()
            }
        }
    }
}

func (synth *Synth) needPushRedis() bool {
    now := time.Now()

    // 以时间间隔为准
    //if now.UnixNano() <= now.Round(synth.pushPeriod).UnixNano() || synth.nextPushTime.IsZero() {
    //    return false
    //} else {
    //    synth.nextPushTime = now.Round(synth.pushPeriod)
    //    return true
    //}

    // 以本地缓存容量为准
    if len(synth.todo) == synth.maxCacheSize {
        synth.nextPushTime = now
        return true
    }
    return false
}

func (synth *Synth) Run() {
    synth.unrelatedExpiredKeysSynth()
    go synth.realTimeSynth()
    go synth.expiredSynth()
}
