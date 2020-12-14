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

type Pipeline struct {
	todoQueue chan (<-chan interface{})
	cmdSocket *runner
}

var (
	redisPipeline        Pipeline
	emptyDelayHandleUnit = DelayHandleUnit{Func: nil, Args: nil}
)

func (rp *Pipeline) asyncStore(k, v string) {

	c := command{name: "SET", args: []interface{}{k, v}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	// 不用处理结果
	//rp.todoQueue <- c.todo
}

func (rp *Pipeline) asyncStoreWithExpire(k, v string, expire int) {

	c := command{name: "SETEX", args: []interface{}{k, expire, v}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	// 不用处理结果
	//rp.todoQueue <- c.todo
}

func (rp *Pipeline) asyncLoad(k string, u DelayHandleUnit) {
	c := command{name: "GET", args: []interface{}{k}, todo: make(chan interface{}, 2)}

	// 1、延迟处理单元先放入通道，2、查询结果后放入
	c.todo <- u
	// 发送没命令
	rp.cmdSocket.send <- c
	// 处理结果以item为单元
	rp.todoQueue <- c.todo
}

func (rp *Pipeline) asyncCollectResult(doResultFunc func(unit DelayHandleUnit, result CmdResult)) {
	for r := range rp.todoQueue {
		// 异步处理redis查询结果 model.HandleRedisResult
		dhu := <-r
		res := <-r

		doResultFunc(dhu.(DelayHandleUnit), res.(CmdResult))
	}
}

func (rp *Pipeline) asyncDelete(k string) {
	c := command{name: "DEL", args: []interface{}{k}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	// 不用处理结果
	//rp.todoQueue <- c.todo
}

func (rp *Pipeline) asyncFlushAll() {
	c := command{name: "FLUSHALL", args: []interface{}{}, todo: make(chan interface{}, 2)}
	c.todo <- emptyDelayHandleUnit
	rp.cmdSocket.send <- c
	// 不用处理结果
	//rp.todoQueue <- c.todo
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

func AsyncFlushAll() {
	redisPipeline.asyncFlushAll()
}

func Init(c *Config, fun func(unit DelayHandleUnit, result CmdResult)) error {

	redisPipeline = Pipeline{
		todoQueue: make(chan (<-chan interface{}), 1<<20),
		cmdSocket: newRunner(newPool(c).Get()),
	}

	go redisPipeline.asyncCollectResult(fun)

	AsyncFlushAll()
	time.Sleep(time.Second * 3)

	return nil
}
