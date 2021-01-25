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

func (d *Dao) CacheSipItems(data []*model.SipItem) {
	var (
		conn = d.redis.Get(context.TODO())
		args = redis.Args{}
		err  error
		bs   []byte
	)
	defer conn.Close()

	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Errorf("json.Marshal err(%v)", err)
			continue
		}
		args = args.Add(v.CallId).Add(string(bs))
	}
	if err = conn.Send("MSET", args...); err != nil {
		log.Errorf("CacheSipItems conn.Send(MSET) error(%v)", err)
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

	return item, nil
}

func (d *Dao) GetSipItems(ids []string) (res []*model.SipItem, err error) {
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
			log.Errorf("GetSipItems conn.Do(MGET) error(%v)", err)
		}
		return
	}

	for _, bs := range bss {
		if bs == nil {
			res = append(res, nil)
			continue
		}

		item := model.NewSipItem()
		if err = json.Unmarshal(bs, item); err != nil {
			item.Free()
			err = nil
			log.Errorf("json.Unmarshal error(%v)", err)
			continue
		}

		res = append(res, item)
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
		args = redis.Args{}
		conn = d.redis.Get(context.TODO())
	)
	defer conn.Close()

	args = args.Add(key)
	for _, v := range ids {
		args = args.Add(v)
	}
	_, err := conn.Do("SADD", args...)
	if err != nil {
		log.Error("conn.Do error(%v)", err)
	}
}

func (d *Dao) DelExpiredSipItemKeysSets(ids []string) {
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
		log.Error("conn.Do(Del) error(%v)", err)
	}
}

func (d *Dao) GetExpiredSipItemKeysSetMembers(k string) (mem []string, err error) {
	var (
		conn = d.redis.Get(context.TODO())
	)
	defer conn.Close()

	if mem, err = redis.Strings(conn.Do("SMEMBERS", k)); err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Errorf("conn.Do(SMEMBERS) error(%v)", err)
		return nil, err
	}
	return
}

func (d *Dao) GetExpiredSipItemKeysSetSize(key string) int {
	var (
		n    int
		err  error
		conn = d.redis.Get(context.TODO())
	)
	defer conn.Close()

	if n, err = redis.Int(conn.Do("SCARD", key)); err != nil {
		log.Error("conn.Do(SCARD) error(%v)", err)
		return 0
	}
	return n
}
