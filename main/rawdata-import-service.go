package main

import (
	"centnet-cdrrs/adapter/kafka"
	"centnet-cdrrs/conf"
	"centnet-cdrrs/library/log"
	"flag"
	"fmt"
	"github.com/satori/go.uuid"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var pc = kafka.ProducerConfig{
	Topic:      "SipPacket",
	Broker:     "192.168.1.205:9092",
	Frequency:  500,
	MaxMessage: 1 << 20,
}

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

	producer, err := kafka.NewProducer(&pc)
	if err != nil {
		panic(err)
	}

	producer.Run()

	mock(producer)

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}

func mock(producer *kafka.Producer) {
	var inviteMessage = `SIP/2.0 200 OK\0D\0AFrom: ""1101385""<sip:1101385@192.168.6.24;user=phone>;tag=04ab01e2d14221470\0D\0ATo: <sip:a52302813007510@219.143.187.139;user=phone>;tag=b0520e20-0-13e0-67e6bb-76fea845-67e6bb\0D\0ACall-ID: 04ab01e2d142787@192.168.6.24\0D\0A[Generated Call-ID: 04ab01e2d142787@192.168.6.24]\0D\0ACSeq: 17 INVITE\0D\0AAllow: INVITE,ACK,OPTIONS,REGISTER,INFO,BYE,UPDATE\0D\0AUser-Agent: DonJin SIP Server 3.2.0_i\0D\0ASupported: 100rel\0D\0AVia: SIP/2.0/UDP 192.168.6.24:5060;received=116.24.65.63;rport=5060;branch=z9hG4bK31596\0D\0AContact: <sip:018926798345@219.143.187.139;user=phone>\0D\0AContent-Type: application/sdp\0D\0AContent-Length: 126\0D\0A\0D\0Av=0\0D\0Ao=1101385 20000001 3 IN IP4 192.168.6.24\0D\0As=A call\0D\0Ac=IN IP4 192.168.6.24\0D\0At=0 0\0D\0Am=audio 10000 RTP/AVP 8 0\0D\0Aa=rtpmap:8 PCMA/8000\0D\0Aa=rtpmap:0 PCMU/8000\0D\0Aa=ptime:20\0D\0Aa=sendrecv\0D\0A`
	var rawInviteMessage = `"20201123120800","10020044201","220.248.118.20","9080","61.220.35.200","8080",` + "\"" + inviteMessage + "\""
	var byeMessage = `SIP/2.0 200 OK\0D\0AFrom: ""1101385""<sip:1101385@192.168.6.24;user=phone>;tag=04ab01e2d14221470\0D\0ATo: <sip:a52302813007510@219.143.187.139;user=phone>;tag=b0520e20-0-13e0-67e6bb-76fea845-67e6bb\0D\0ACall-ID: 04ab01e2d142787@192.168.6.24\0D\0A[Generated Call-ID: 04ab01e2d142787@192.168.6.24]\0D\0ACSeq: 18 BYE\0D\0AVia: SIP/2.0/UDP 192.168.6.24:5060;received=116.24.65.63;rport=5060;branch=z9hG4bK9701\0D\0ASupported: 100rel\0D\0AContent-Length: 0\0D\0A`
	var rawByeMessage = `"20201123120805","10020044201","220.248.118.20","9080","61.220.35.200","8080",` + "\"" + byeMessage + "\""

	const maxSize = 1
	var uuidList [maxSize]string
	for i := 0; i < maxSize; i++ {
		uuidList[i] = uuid.NewV4().String()
	}

	t := time.Now()
	for i := 0; i < maxSize; i++ {
		ss := strings.Split(rawInviteMessage, `\0D\0A`)
		for x, s := range ss {
			if strings.HasPrefix(s, "Call-ID: ") {
				ss[x] = "Call-ID: " + uuidList[i]
			}
		}
		rawInviteMessage = strings.Join(ss, `\0D\0A`)
		producer.Log(uuidList[i], rawInviteMessage)
	}
	log.Debugf("INVITE-200OK messages import completed. %d packets, %v, %.f pps", maxSize, time.Since(t), maxSize/time.Since(t).Seconds())

	time.Sleep(time.Second * 5)
	log.Debug("5 seconds sleeping...")
	t = time.Now()
	for i := 0; i < maxSize; i++ {
		ss := strings.Split(rawByeMessage, `\0D\0A`)
		for x, s := range ss {
			if strings.HasPrefix(s, "Call-ID: ") {
				ss[x] = "Call-ID: " + uuidList[i]
			}
		}
		rawByeMessage = strings.Join(ss, `\0D\0A`)
		producer.Log(uuidList[i], rawByeMessage)
	}
	log.Debugf("BYE-200OK messages import completed. %d packets, %v, %.f pps", maxSize, time.Since(t), maxSize/time.Since(t).Seconds())
	fmt.Println("Log Over")
}
