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
	if err = redis.Init(conf.Conf.Redis, model.HandleRedisResult); err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	/* 数据库模块初始化 */
	err = dao.Init(conf.Conf.Mysql)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	/* 开启解析 */
	file.NewRawFileParser(conf.Conf.FileParser).Run(model.DoLine)

	/* 还原的话单数据交给诈骗分析模型 */
	//fraudAnalysisProducer, err := kafka.NewProducer(conf.Conf.Kafka.FraudModelProducer)
	//if err != nil {
	//	log.Error(err)
	//	os.Exit(-1)
	//}
	//fraudAnalysisProducer.Run()
	//
	//c := conf.Conf.Kafka.RestoreCDRConsumer
	//for i := 1; i <= c.GroupMembers; i++ {
	//	clientID := fmt.Sprintf("member_%02d", i)
	//	sipPacketConsumer := kafka.NewConsumerGroupMember(c, clientID, model.RestoreCDR)
	//	if sipPacketConsumer == nil {
	//		log.Error("kafka.NewConsumerGroupMember Error")
	//		os.Exit(-1)
	//	}
	//	sipPacketConsumer.SetNextPipeline(fraudAnalysisProducer)
	//}

	//mock(fraudAnalysisProducer)

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}
