package service

import (
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/model"
	"centnet-cdrrs/service/cdr"
	"time"
)

// 话单合成器

var (
	_keysExpirationPeriod = time.Hour
	_setIdsTimeFormat     = "20060102150405"
	_defaultExpiredFunc   = func(interface{}) {}
)

type Synth struct {
	todo           []*model.SipItem
	ch             chan *model.SipItem
	dao            *dao.Dao
	cdrPro         *cdr.CDRProducer
	onExpired      func(interface{})
	nextPushTime   time.Time // 下次推送redis时间
	oldestTodoTime time.Time // 最旧待处理key时间
	pushPeriod     time.Duration
}

func newSynthesizer() *Synth {
	return &Synth{
		ch:         make(chan *model.SipItem),
		onExpired:  _defaultExpiredFunc,
		pushPeriod: time.Second * 2,
	}
}

func (synth *Synth) Input(item *model.SipItem) {
	synth.ch <- item
}

func (synth *Synth) expiredSetIds(now time.Time) (ids []string) {
	var ss []string
	for i := 1; ; i++ {
		if now.Sub(synth.oldestTodoTime) >= _keysExpirationPeriod*time.Duration(i) {
			t := now.Add(-_keysExpirationPeriod * time.Duration(i))
			ss = append(ss, t.Format(_setIdsTimeFormat))
			continue
		}
		break
	}
	if len(ss) == 0 {
		return
	}

	// 按时间顺序排序
	for i := len(ss); i > 0; i-- {
		ids = append(ids, ss[i-1])
	}

	return
}

func (synth *Synth) OnExpired(handle func(interface{})) {
	synth.onExpired = handle
}

// 定时处理超时key并输出话单
func (synth *Synth) expiredSynth() {
	var (
		err    error
		keys   []string
		items  []*model.SipItem
		ticker = time.NewTicker(time.Second)
	)

	for {
		select {
		case now := <-ticker.C:
			// 没有超时的键
			if synth.oldestTodoTime.IsZero() || now.Sub(synth.oldestTodoTime) < _keysExpirationPeriod {
				continue
			}

			for _, sid := range synth.expiredSetIds(now) {
				// 获取超时集合里所有key
				keys, err = synth.dao.GetExpiredSipItemKeysSetMembers(sid)
				if err != nil || len(keys) == 0 {
					continue
				}

				// 获取所有key值
				items, err = synth.dao.GetSipItems(keys)
				if err != nil || len(items) == 0 {
					continue
				}

				// TODO: 输出未成功合成的话单
				synth.onExpired(items)

				// 清除redis缓存
				synth.dao.DelSipItems(keys)
				synth.dao.DelExpiredSipItemKeysSets([]string{sid})
			}
		}
	}
}

func (synth *Synth) sipItemKeys() (ks []string) {
	for _, v := range synth.todo {
		ks = append(ks, v.CallId)
	}
	return
}

func (synth *Synth) sipItemKeysSetId() string {
	return synth.nextPushTime.Format(_setIdsTimeFormat)
}

func (synth *Synth) clearSipItems() {
	for _, v := range synth.todo {
		if v != nil {
			v.Free()
		}
	}

	synth.todo = synth.todo[:0]
}

func (synth *Synth) trySynthesize() {

	// 批量查询各报文对应的200OK包
	allQueried, err := synth.dao.GetSipItems(synth.sipItemKeys())
	if err != nil {
		log.Error("synth.dao.GetSipItems error(%v)", err)
		return
	}

	// 依次处理查询结果
	var queriedKeys []string
	var notQueried []*model.SipItem
	for i, v := range synth.todo {
		// 未查询到集中缓存
		if allQueried[i] == nil {
			notQueried = append(notQueried, v)
			continue
		}

		// 查询到立即合成话单
		if v.Type == model.SipStatusInvite200OK && allQueried[i].Type == model.SipStatusBye200OK {
			v.DisconnectTime = allQueried[i].DisconnectTime
			if c := synth.cdrPro.Gen(v.CallId, v); c != nil {
				synth.cdrPro.Put(c)
			}
			queriedKeys = append(queriedKeys, allQueried[i].CallId)
		} else if v.Type == model.SipStatusBye200OK && allQueried[i].Type == model.SipStatusInvite200OK {
			allQueried[i].DisconnectTime = v.DisconnectTime
			if c := synth.cdrPro.Gen(allQueried[i].CallId, allQueried[i]); c != nil {
				synth.cdrPro.Put(c)
			}
			queriedKeys = append(queriedKeys, allQueried[i].CallId)
		}
		v.Free()
		allQueried[i].Free()
	}

	synth.dao.DelSipItems(queriedKeys)
	synth.todo = notQueried
}

// 合成话单
func (synth *Synth) realTimeSynth() {
	var err error

	for {
		select {
		case item := <-synth.ch:
			// 计算当前数据所属set id
			nextTime := nextPushTime(synth.pushPeriod)
			if synth.nextPushTime == nextTime {
				synth.todo = append(synth.todo, item)
				continue
			}

			// 尝试实时合成
			synth.trySynthesize()

			// 不是第一次计算set id，处理缓存
			if !synth.nextPushTime.IsZero() {
				// 缓存key-value
				synth.dao.CacheSipItems(synth.todo)

				// 缓存key集合
				sid, ids := synth.sipItemKeysSetId(), synth.sipItemKeys()
				synth.dao.CacheExpiredSipItemKeysSet(sid, ids)

				synth.clearSipItems()
			} else {
				synth.oldestTodoTime = synth.nextPushTime
				if err != nil {
					log.Error(err)
				}
			}

			synth.todo = append(synth.todo, item)
			synth.nextPushTime = nextTime
		}
	}
}

func nextPushTime(period time.Duration) time.Time {
	return time.Now().Round(period)
}

func (synth *Synth) Run() {
	go synth.realTimeSynth()
	go synth.expiredSynth()
}