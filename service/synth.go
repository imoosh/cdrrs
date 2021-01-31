package service

import (
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/model"
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

	todo    map[string]*model.SipCache
	cdrCh   chan *model.SipCache
	exSynth chan exSynth

	dao       *dao.Dao
	onExpired func(interface{})

	pushPeriod   time.Duration // 由时间间隔决定本地缓存容量
	maxCacheSize int           // 由缓存大小决定本地缓存容量
}

// 超时合成话单
type exSynth struct {
	sid   string      // 缓存keys集合id
	timer *time.Timer // 超时处理定时器
}

func newSynthesizer(svr *Service) *Synth {
	return &Synth{
		c:            svr.c,
		todo:         make(map[string]*model.SipCache),
		cdrCh:        make(chan *model.SipCache),
		exSynth:      make(chan exSynth, 10000),
		onExpired:    _defaultExpiredFunc,
		pushPeriod:   time.Millisecond * 500,
		maxCacheSize: 10000,
		dao:          svr.dao,
	}
}

func (synth *Synth) Input(item *model.SipCache) {
	synth.cdrCh <- item
}

func (synth *Synth) OnExpired(handle func(interface{})) {
	synth.onExpired = handle
}

func (synth *Synth) handleExpiredKeySet(sid string) {
	var (
		err   error
		keys  []string
		items map[string]*model.SipCache
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
	return strings.Replace(time.Now().Format(_setIdsTimeFormat), ".", "", 1)
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

func (synth *Synth) synthesize(a, b *model.SipCache) bool {
	if a.Type == model.SipStatusInvite200OK && b.Type == model.SipStatusBye200OK {
		if a.CallId != b.CallId {
			log.Panicf("a.CallId(%s) != b.CallId(%s)", a.CallId, b.CallId)
		}
		a.DisconnectTime = b.DisconnectTime
		if c := model.NewCDR(a.CallId, a); c != nil {
			//synth.cdrPro.Put(c)
			Collect(c)
		}
		return true
	} else if a.Type == model.SipStatusBye200OK && b.Type == model.SipStatusInvite200OK {
		if a.CallId != b.CallId {
			log.Panicf("a.CallId(%s) != b.CallId(%s)", a.CallId, b.CallId)
		}
		b.DisconnectTime = a.DisconnectTime
		if c := model.NewCDR(b.CallId, b); c != nil {
			//synth.cdrPro.Put(c)
			Collect(c)
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

func (synth *Synth) append(item *model.SipCache) {
	// 先尝试再本地缓存中匹对数据
	if val, ok := synth.todo[item.CallId]; ok {
		if val.CallId != item.CallId {
			log.Fatalf("val.CallId(%s) != item.CallId(%s)", val.CallId, item.CallId)
		}

		// 话单合成功，才会是放map中缓存，否则会等待
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

			synth.append(item)
			if synth.needPushRedis() {
				pushRedis()
				ctime = time.Now()
				continue
			}
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
	// 以本地缓存容量满了推送至redis
	if len(synth.todo) == synth.maxCacheSize {
		return true
	}
	return false
}

func (synth *Synth) Run() {
	// 处理系统重启前缓存的话单原始数据
	synth.unrelatedExpiredKeysSynth()

	// 实时合成话单
	go synth.realTimeSynth()

	// 处理超时的缓存话单原始数据
	go synth.expiredSynth()
}
