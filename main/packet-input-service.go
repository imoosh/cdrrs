package main

import (
	"centnet-cdrrs/adapter/kafka/analytic"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var pc = analytic.ProducerConfig{
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
	producer, err := analytic.NewProducer(&pc)
	if err != nil {
		panic(err)
	}

	producer.Run()

	var sipMessage = `INVITE sip:018926798345@219.143.187.139;user=phone SIP/2.0\0D\0AVia: SIP/2.0/UDP 192.168.6.24:5060;branch=z9hG4bK31596;rport\0D\0AFrom: ""1101385"" <sip:1101385@192.168.6.24;user=phone>;tag=04ab01e2d14221470\0D\0ATo: <sip:018926798345@219.143.187.139;user=phone>\0D\0ACall-ID: 04ab01e2d142787@192.168.6.24\0D\0ACSeq: 17 INVITE\0D\0AContact: ""1101385"" <sip:1101385@192.168.6.24:5060>\0D\0AMax-Forwards: 70\0D\0AUser-Agent: SIPUA\0D\0AExpires: 120\0D\0ASupported: 100rel\0D\0AP-Preferred-Identity: ""1101385"" <sip:1101385@192.168.6.24;user=phone>\0D\0AAllow: INVITE, ACK, CANCEL, BYE, OPTIONS, REFER, SUBSCRIBE, NOTIFY, MESSAGE, UPDATE, PRACK\0D\0AContent-Type: application/sdp\0D\0AContent-Length: 182\0D\0A\0D\0Av=0\0D\0Ao=1101385 20000001 3 IN IP4 192.168.6.24\0D\0As=A call\0D\0Ac=IN IP4 192.168.6.24\0D\0At=0 0\0D\0Am=audio 10000 RTP/AVP 8 0\0D\0Aa=rtpmap:8 PCMA/8000\0D\0Aa=rtpmap:0 PCMU/8000\0D\0Aa=ptime:20\0D\0Aa=sendrecv\0D\0A`
	var rawMessage = `"20201111100710","10020044169","220.248.118.20","9080","61.220.35.200","8080",` + "\"" + sipMessage + "\""

	for i := 1; i <= 10; i++ {
		producer.Log(fmt.Sprintf("%08d", i), rawMessage)
	}

	fmt.Println("Log Over")

	// os signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
}
