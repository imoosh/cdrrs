package redis

// https://gist.github.com/garyburd/c3219db8194841adff6f97f5dec28f60

import (
	"centnet-cdrrs/library/log"
	"time"

	"github.com/gomodule/redigo/redis"
)

type command struct {
	name string
	args []interface{}

	todo chan interface{}
	unit DelayHandleUnit
}

type CmdResult struct {
	Err   error
	Value interface{}
}

type runner struct {
	conn redis.Conn
	send chan command
	recv chan chan interface{}
	stop chan struct{}
	done chan struct{}
}

func (r *runner) sender() {
	count := 0
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-r.stop:
			if err := r.conn.Flush(); err != nil {
				log.Error(err)
			}
			close(r.recv)
			return
		case cmd := <-r.send:
			if err := r.conn.Send(cmd.name, cmd.args...); err != nil {
				log.Error(err)
			}
			// Flush if the send queue is empty or the CmdResult queue is full.
			if len(r.send) == 0 || len(r.recv) == cap(r.recv) {
				if err := r.conn.Flush(); err != nil {
					log.Error(err)
				}
			}
			r.recv <- cmd.todo
			count++
			if count%10000 == 0 {
				log.Debug("send command count:", count)
			}
		case <-ticker.C:
			if err := r.conn.Flush(); err != nil {
				log.Error(err)
			}
		}
	}
}

func (r *runner) receiver() {

	for ch := range r.recv {
		var result CmdResult
		result.Value, result.Err = r.conn.Receive()
		// 将结果放入通道中
		ch <- result
		// 发送完立即关闭
		close(ch)
		if result.Err != nil && r.conn.Err() != nil {
			log.Error(r.conn.Err())
		}
	}
	close(r.done)
}

func newRunner(conn redis.Conn) *runner {
	r := &runner{
		conn: conn,
		send: make(chan command, 4096),
		recv: make(chan chan interface{}, 4096),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go r.sender()
	go r.receiver()
	return r
}

//func main() {
//	totalRequests := flag.Int("n", 100000, "Total number of `requests`")
//	flag.Parse()
//	r := newRunner()

//
//	start := time.Now()
//	args := []interface{}{"a", "b"}
//	for i := 0; i < *totalRequests; i++ {
//		r.send <- command{name: "SET", args: args, CmdResult: make(chan CmdResult, 1)}
//	}
//	close(r.stop)
//	<-r.done
//	t := time.Since(start)
//	fmt.Printf("%d requests completed in %f seconds\n", *totalRequests, float64(t)/float64(time.Second))
//	fmt.Printf("%f requests / second\n", float64(*totalRequests)*float64(time.Second)/float64(time.Since(start)))
//}
