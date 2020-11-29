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
	c := conf.Conf.Kafka.SipPacketConsumer
	for i := 1; i <= c.GroupMembers; i++ {
		clientID := fmt.Sprintf("GroupClient_%02d", i)
		sipPacketConsumer := kafka.NewConsumerGroupMember(c, clientID, model.AnalyzePacket)
		if sipPacketConsumer == nil {
			os.Exit(-1)
		}
		sipPacketConsumer.SetNextPipeline(restoreCDRProducer)
	}

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}
