package redis

import (
	"centnet-cdrrs/conf"
	"fmt"
	"testing"
)

func TestConn_Put(t *testing.T) {
	redisConn.Put("04ab01e2d142787@192.168.6.24", "{}")
	fmt.Println("insert redis ok")
}

func TestConn_PutWithExpire(t *testing.T) {
	redisConn.PutWithExpire("04ab01e2d142787@192.168.6.24", "{}", conf.Conf.Redis.CacheExpire)
}

func TestConn_Delete(t *testing.T) {
	redisConn.Delete("04ab01e2d142787@192.168.6.24")
}
