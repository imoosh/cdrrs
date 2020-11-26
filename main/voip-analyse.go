package main

import (
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/model"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	runtime.GOMAXPROCS(runtime.NumCPU())

	/* 解析参数 */
	flag.Parse()
	conf.Init()
	fmt.Println(conf.Conf)

	/* 日志模块初始化 */
	log.Init(conf.Conf.Logging)

	/* 数据库模块初始化 */
	err := dao.Init(conf.Conf.Mysql)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	/* 待还原话单数据生产者 */
	restoreCDRProducer, err := kafka.NewProducer(conf.Conf.Kafka.RestoreCDRProducer)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	restoreCDRProducer.Run()

	/* sip包数据消费者 */
	sipPacketConsumer := kafka.NewConsumer(conf.Conf.Kafka.SipPacketConsumer, model.AnalyzePacket)
	if sipPacketConsumer == nil {
		log.Error("NewConsumer Error.")
		os.Exit(-1)
	}
	/* 解析完的sip包数据交给下一级的生产者处理 */
	sipPacketConsumer.SetNextProducer(restoreCDRProducer)
	err = sipPacketConsumer.Run()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}
