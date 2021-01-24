package service

import (
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/model"
	"time"
)

// 话单合成器

var (
	KeysDuration = time.Hour
)

type synthesizer struct {
	ch          chan *model.SipItem
	dao         *dao.Dao
	oldestSetId string
	lastSetId   string
}

func newSynthesizer() *synthesizer {
	return &synthesizer{
		ch: make(chan *model.SipItem),
	}
}

func (s *synthesizer) Input(item *model.SipItem) {
	s.ch <- item
}

func expiredKeysSetId(period time.Duration) string {
	return time.Now().Truncate(period).Format("20060102150405")
}

func (s *synthesizer) expiredDeleter() {
	var (
		err        error
		oldestTime time.Time
		items      map[string]*model.SipItem
		ticker     = time.NewTicker(time.Second)
	)

	go func() {
		for {
			select {
			case <-ticker.C:
				if len(s.oldestSetId) == 0 {
					continue
				}

				if (oldestTime == time.Time{}) {
					oldestTime, err = time.Parse("20060102150405", s.oldestSetId)
					if err != nil {
						log.Error(err)
						continue
					}
				}

				// 键已超时，需要通过id+1计算超时id，因ticker有误差，不能通过ticker计算
				now := time.Now()
				if now.Sub(oldestTime) >= KeysDuration {
					keysId := now.Add(-KeysDuration).Format("20060102150405")
					keys, err := s.dao.GetExpiredSipItemKeysSetMembers(keysId)
					if err != nil || len(keys) == 0 {
						continue
					}

					items, err = s.dao.GetSipItems(keys)
					if err != nil || len(items) == 0 {
						continue
					}

					// TODO: 输出未成功合成的话单
					// ...

					// 清除redis缓存
					s.dao.DelSipItems(keys)
					s.dao.DelExpiredSipItemKeysSets([]string{keysId})
				}

			}
		}
	}()
}

func (s *synthesizer) daemon() {
	var ids []string

	var siMap map[string]*model.SipItem

	siMap = make(map[string]*model.SipItem)

	go func() {
		for {
			select {
			case item := <-s.ch:
				keysSetId := expiredKeysSetId(time.Second)
				if s.lastSetId == keysSetId {
					siMap[item.CallId] = item
					continue
				}

				if len(s.lastSetId) != 0 {
					s.dao.CacheSipItems(siMap)
					for k, v := range siMap {
						ids = append(ids, v.CallId)
						delete(siMap, k)
						v.Free()
					}
					s.dao.CacheExpiredSipItemKeysSet(s.lastSetId, ids)
					ids = ids[:0]
				} else {
					s.oldestSetId = keysSetId
				}

				siMap[item.CallId] = item
				s.lastSetId = keysSetId
			}
		}
	}()
}
