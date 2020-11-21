package redis

import (
	"centnet-cdrrs/conf"
	"testing"
)

func TestConn_Put(t *testing.T) {
	RedisConn.Put("04ab01e2d142787@192.168.6.24", "{}")
}

func TestConn_PutWithExpire(t *testing.T) {
	RedisConn.PutWithExpire("04ab01e2d142787@192.168.6.24", "{}", conf.Conf.Redis.CacheExpire)
}

func TestConn_Delete(t *testing.T) {
	RedisConn.Delete("04ab01e2d142787@192.168.6.24")
}
