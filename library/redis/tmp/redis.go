package tmp

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

var rdb *redis.Client

func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "192.168.1.205:6379",
		Password: "",
		DB:       0,
	})
	_, err = rdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func Test() {

}

func TxPipeline() {
	pipe := rdb.TxPipeline()
	pipe.Set("name", "wayne", time.Minute*30)
	pipe.Get("name")
	pipe.Del("name")

	cmds, err := pipe.Exec()
	if err != nil {
		fmt.Println(err)
	}

	for _, cmd := range cmds {
		fmt.Println(cmd.Name())
	}
}
