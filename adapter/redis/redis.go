package redis

import (
	"time"
)

type Config struct {
	Host        string
	Password    string
	IdleTimeout int
	MaxIdle     int
	MaxActive   int
	CacheExpire int
}

type DelayHandleUnit struct {
	Func func(interface{}, interface{}) interface{}
	Args interface{}
}

type collection chan (<-chan interface{})

type RedisPipeline struct {
	todoQueue collection
	cmdSocket *runner

	ResultFunc func(unit DelayHandleUnit, result CmdResult)
}

var (
	redisPipeline        RedisPipeline
	emptyDelayHandleUnit = DelayHandleUnit{Func: nil, Args: nil}
)

func (rp *RedisPipeline) asyncStore(k, v string) {

	c := command{name: "SET", args: []interface{}{k, v}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	rp.todoQueue <- c.todo
}

func (rp *RedisPipeline) asyncStoreWithExpire(k, v string, expire int) {

	c := command{name: "SETEX", args: []interface{}{k, expire, v}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	rp.todoQueue <- c.todo
}

func (rp *RedisPipeline) asyncLoad(k string, u DelayHandleUnit) {
	c := command{name: "GET", args: []interface{}{k}, todo: make(chan interface{}, 2)}

	// 1、延迟处理单元先放入通道，2、查询结果后放入
	c.todo <- u
	// 发送没命令
	rp.cmdSocket.send <- c
	// 处理结果以item为单元
	rp.todoQueue <- c.todo
}

func (rp *RedisPipeline) asyncDelete(k string) {
	c := command{name: "DEL", args: []interface{}{k}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	rp.todoQueue <- c.todo
}

func (rp *RedisPipeline) asyncCollectResult() {
	for r := range rp.todoQueue {
		// 异步处理redis查询结果 model.HandleRedisResult
		rp.ResultFunc((<-r).(DelayHandleUnit), (<-r).(CmdResult))
	}
}

func AsyncStore(k, v string) {
	redisPipeline.asyncStore(k, v)
}

func AsyncStoreWithExpire(k, v string, expire time.Duration) {
	redisPipeline.asyncStoreWithExpire(k, v, int(expire.Seconds()))
}

func AsyncLoad(k string, unit DelayHandleUnit) {
	redisPipeline.asyncLoad(k, unit)
}

func AsyncDelete(k string) {
	redisPipeline.asyncDelete(k)
}

func Init(c *Config, fun func(unit DelayHandleUnit, result CmdResult)) error {

	redisPipeline = RedisPipeline{
		todoQueue:  make(collection, 10000),
		cmdSocket:  newRunner(newPool(c).Get()),
		ResultFunc: fun,
	}

	go redisPipeline.asyncCollectResult()

	return nil
}
