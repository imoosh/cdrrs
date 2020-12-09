package main

import (
	"centnet-cdrrs/adapter/file"
	"centnet-cdrrs/adapter/redis"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model"

	//"encoding/json"
	"flag"
	"fmt"
	//"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	//"time"
)

var (
	showVersion bool
	BuiltID     string
	BuiltHost   string
	BuiltTime   string
	GoVersion   string
)

func init() {
	flag.BoolVar(&showVersion, "v", false, "show application version and exit")

	if !flag.Parsed() {
		flag.Parse()
	}

	if showVersion {
		fmt.Println(getAppVersion())
		os.Exit(0)
	}
}

func getAppVersion() string {
	return fmt.Sprintf(""+
		"Built ID:   %s\n"+
		"Built Host: %s\n"+
		"Built Time: %s\n"+
		"Go Vesrion: %s\n",
		BuiltID, BuiltHost, BuiltTime, GoVersion)
}

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())

	/* 解析参数 */
	flag.Parse()
	conf.Init()
	fmt.Println(conf.Conf)

	/* 日志模块初始化 */
	log.Init(conf.Conf.Logging)

	/* redis初始化 */
	if err = redis.Init(conf.Conf.Redis, model.DoRedisResult); err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	/* 数据库模块初始化 */
	err = dao.Init(conf.Conf.Mysql)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	err = model.Init()
	if err != nil {
		log.Error("model.Init failed")
		os.Exit(-1)
	}

	/* 开启文件解析 */
	file.NewRawFileParser(conf.Conf.FileParser).Run(model.DoLine)

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}
