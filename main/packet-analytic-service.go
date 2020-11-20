package main

import (
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/dao"
	"centnet-cdrrs/library/log"
	"centnet-cdrrs/prot/sip"
	"flag"
	"fmt"
	"github.com/astaxie/beego/orm"
	"os"
	"os/signal"
	"runtime"
	"sync"
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
	//err := dao.Init(conf.Conf.Mysql)
	//if err != nil {
	//    log.Error(err)
	//    os.Exit(-1)
	//}

	/* 待还原话单数据生产者 */
	restoreCDRProducer, err := kafka.NewProducer(conf.Conf.Kafka.RestoreCDRProducer)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	restoreCDRProducer.Run()

	/* sip包数据生产者 */
	sipPacketProducer, err := kafka.NewProducer(conf.Conf.Kafka.SipPacketProducer)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	sipPacketProducer.Run()

	/* sip包数据消费者 */
	sipPacketConsumer := kafka.NewConsumer(conf.Conf.Kafka.SipPacketConsumer, kafka.AnalyzePacket)
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

func main2() {

	sipPacket := dao.Sip{
		Sip:           "192.168.1.98",
		Sport:         5060,
		Dip:           "192.168.1.14",
		Dport:         5060,
		CallId:        "b5deab6380c4e57fa20486e493c68324",
		ReqMethod:     "INVITE",
		ReqStatusCode: 5060,
		ReqUser:       "wayne",
		ReqHost:       "dvao.cn",
		ReqPort:       5060,
		FromName:      "wayne",
		FromUser:      "wayne",
		FromHost:      "dvao.cn",
		FromPort:      5060,
		ToName:        "abc",
		ToUser:        "abc",
		ToHost:        "asdkl;jf.ca",
		ToPort:        5060,
		ContactName:   "abc",
		ContactUser:   "abc",
		ContactHost:   "ak;lfa.cn",
		ContactPort:   5060,
		CseqMethod:    "INVITE",
		UserAgent:     "cnx3000",
	}

	o := orm.NewOrm()
	n, err := o.Insert(&sipPacket)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(n)
}

func main1() {
	// Load up a test message
	raw := []byte("SIP/2.0 200 OK\r\n" +
		"Via: SIP/2.0/UDP 192.168.2.242:5060;received=22.23.24.25;branch=z9hG4bK5ea22bdd74d079b9;alias;rport=5060\r\n" +
		"To: <sip:JohnSmith@mycompany.com>;tag=aprqu3hicnhaiag03-2s7kdq2000ob4\r\n" +
		"From: sip:HarryJones@mycompany.com;tag=89ddf2f1700666f272fb861443003888\r\n" +
		"CSeq: 57413 REGISTER\r\n" +
		"Call-ID: b5deab6380c4e57fa20486e493c68324\r\n" +
		"Contact: <sip:JohnSmith@192.168.2.242:5060>;expires=192\r\n\r\n")

	var wg sync.WaitGroup
	wg.Add(6)

	for a := 0; a < 6; a++ {
		go func() {
			for i := 0; i < 1; i++ {
				sip.Parse(raw)
				sip := sip.Parse(raw)
				fmt.Println("From: ", string(sip.From.User), " To: ", string(sip.To.User))
				fmt.Println(string(sip.Contact.User), string(sip.Req.User))
				fmt.Println(string(sip.Req.StatusCode), string(sip.Req.Method), string(sip.Cseq.Method))
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
