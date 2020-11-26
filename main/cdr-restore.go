package main

import (
	"centnet-cdrrs/adapter/kafka"
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
	if err = redis.InitRedisPool(conf.Conf.Redis); err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	/* 数据库模块初始化 */
	err = dao.Init(conf.Conf.Mysql)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	/* 还原的话单数据交给诈骗分析模型 */
	fraudAnalysisProducer, err := kafka.NewProducer(conf.Conf.Kafka.FraudModelProducer)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	fraudAnalysisProducer.Run()

	/* sip包数据消费者 */
	restoreCDRConsumer := kafka.NewConsumer(conf.Conf.Kafka.RestoreCDRConsumer, model.RestoreCDR)
	if restoreCDRConsumer == nil {
		log.Error("NewConsumer Error.")
		os.Exit(-1)
	}
	/* 解析完的sip包数据交给下一级的生产者处理 */
	restoreCDRConsumer.SetNextProducer(fraudAnalysisProducer)
	err = restoreCDRConsumer.Run()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	//mock(fraudAnalysisProducer)

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}

//func mock(producer *kafka.Producer) {
//
//	rand.Seed(time.Now().UnixNano())
//
//	for i := 0; i < 1000; i++ {
//		cdr := dao.VoipRestoredCdr{
//			Id:             int64(i),
//			CallId:         fmt.Sprintf("04ab01e2d142787@192.168.6.24-%04d", i),
//			CallerIp:       fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)),
//			CallerPort:     rand.Intn(65535),
//			CalleeIp:       fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)),
//			CalleePort:     rand.Intn(65535),
//			CallerNum:      fmt.Sprintf("110%04d", i/10000),
//			CalleeNum:      fmt.Sprintf("1892679%04d", i),
//			CallerDevice:   "SIPUA",
//			CalleeDevice:   "DonJin SIP Server 3.2.0_i",
//			CalleeProvince: "四川省",
//			CalleeCity:     "成都市",
//			ConnectTime:    time.Now(),
//			DisconnectTime: time.Now(),
//			Duration:       0,
//			FraudType:      "",
//		}
//		cdr.Duration = int(cdr.DisconnectTime.Sub(cdr.ConnectTime).Seconds())
//
//		jsonStr, _ := json.Marshal(cdr)
//		log.Debug(string(jsonStr))
//		producer.Log("cdr", string(jsonStr))
//	}
//
//	for i := 1000; i < 2000; i++ {
//		cdr := dao.VoipRestoredCdr{
//			Id:             int64(i),
//			CallId:         fmt.Sprintf("04ab01e2d142787@192.168.6.24-%04d", i),
//			CallerIp:       fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)),
//			CallerPort:     rand.Intn(65535),
//			CalleeIp:       fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)),
//			CalleePort:     rand.Intn(65535),
//			CallerNum:      fmt.Sprintf("110%04d", i/10000),
//			CalleeNum:      fmt.Sprintf("1892679%04d", i),
//			CallerDevice:   "SIPUA",
//			CalleeDevice:   "DonJin SIP Server 3.2.0_i",
//			CalleeProvince: "四川省",
//			CalleeCity:     "成都市",
//			ConnectTime:    time.Now().Add(-time.Duration(rand.Intn(7) * 1e9)),
//			DisconnectTime: time.Now().Add(time.Duration(rand.Intn(7) * 1e9)),
//			Duration:       0,
//			FraudType:      "",
//		}
//		cdr.Duration = int(cdr.DisconnectTime.Sub(cdr.ConnectTime).Seconds())
//
//		jsonStr, _ := json.Marshal(cdr)
//		log.Debug(string(jsonStr))
//		producer.Log("cdr", string(jsonStr))
//	}
//	//dao.InsertCDR(cdr)
//}
