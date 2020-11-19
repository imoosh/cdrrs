package analytic

import (
	"centnet-cdrrs/library/log"
	"fmt"
	"github.com/Shopify/sarama"
	"strings"
	"sync"
	"time"
)

var pc = ProducerConfig{
	Topic:      "SipPacket",
	Broker:     "192.168.1.205:9092",
	Frequency:  500,
	MaxMessage: 1 << 20,
}

// Config 配置
type ProducerConfig struct {
	Topic      string `xml:"topic"`
	Broker     string `xml:"broker"`
	Frequency  int    `xml:"frequency"`
	MaxMessage int    `xml:"max_message"`
}

type Producer struct {
	producer sarama.AsyncProducer

	topic     string
	msgQ      chan *sarama.ProducerMessage
	wg        sync.WaitGroup
	closeChan chan struct{}
}

// NewProducer 构造KafkaProducer
func NewProducer(cfg *ProducerConfig) (*Producer, error) {

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse                                  // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy                            // Compress messages
	config.Producer.Flush.Frequency = time.Duration(cfg.Frequency) * time.Millisecond // Flush batches every 500ms
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	p, err := sarama.NewAsyncProducer(strings.Split(cfg.Broker, ","), config)
	if err != nil {
		return nil, err
	}
	ret := &Producer{
		producer:  p,
		topic:     cfg.Topic,
		msgQ:      make(chan *sarama.ProducerMessage, cfg.MaxMessage),
		closeChan: make(chan struct{}),
	}

	return ret, nil
}

func (p *Producer) Run() {
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()

	LOOP:
		for {
			select {
			case m := <-p.msgQ:
				p.producer.Input() <- m
			case err := <-p.producer.Errors():
				if err != nil && err.Msg != nil {
					fmt.Printf("[producer] err=[%s] topic=[%s] key=[%s] val=[%s]\n",
						err.Error(), err.Msg.Topic, err.Msg.Key, err.Msg.Value)
				}
			case <-p.closeChan:
				break LOOP
			}
		}
	}()

	for hasTask := true; hasTask; {
		select {
		case m := <-p.msgQ:
			p.producer.Input() <- m
		default:
			hasTask = false
		}
	}
}

func (p *Producer) Close() error {
	close(p.closeChan)
	fmt.Println("producer quit now")
	p.wg.Wait()
	fmt.Println("producer quit ok")

	return p.producer.Close()
}

func (p *Producer) Log(key, value string) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	select {
	case p.msgQ <- msg:
		return
	default:
		log.Debug("[producer] err=[msgQ is full] key=[%s] val=[%s]", msg.Key, msg.Value)
	}
}

//func main() {
//  producer, err := NewProducer(&pc)
//  if err != nil {
//      panic(err)
//  }
//
//  producer.Run()
//
//  var sipMessage = `INVITE sip:018926798345@219.143.187.139;user=phone SIP/2.0\0D\0AVia: SIP/2.0/UDP 192.168.6.24:5060;branch=z9hG4bK31596;rport\0D\0AFrom: ""1101385"" <sip:1101385@192.168.6.24;user=phone>;tag=04ab01e2d14221470\0D\0ATo: <sip:018926798345@219.143.187.139;user=phone>\0D\0ACall-ID: 04ab01e2d142787@192.168.6.24\0D\0ACSeq: 17 INVITE\0D\0AContact: ""1101385"" <sip:1101385@192.168.6.24:5060>\0D\0AMax-Forwards: 70\0D\0AUser-Agent: SIPUA\0D\0AExpires: 120\0D\0ASupported: 100rel\0D\0AP-Preferred-Identity: ""1101385"" <sip:1101385@192.168.6.24;user=phone>\0D\0AAllow: INVITE, ACK, CANCEL, BYE, OPTIONS, REFER, SUBSCRIBE, NOTIFY, MESSAGE, UPDATE, PRACK\0D\0AContent-Type: application/sdp\0D\0AContent-Length: 182\0D\0A\0D\0Av=0\0D\0Ao=1101385 20000001 3 IN IP4 192.168.6.24\0D\0As=A call\0D\0Ac=IN IP4 192.168.6.24\0D\0At=0 0\0D\0Am=audio 10000 RTP/AVP 8 0\0D\0Aa=rtpmap:8 PCMA/8000\0D\0Aa=rtpmap:0 PCMU/8000\0D\0Aa=ptime:20\0D\0Aa=sendrecv\0D\0A`
//  var rawMessage = `"20201111100710","10020044169","220.248.118.20","9080","61.220.35.200","8080",` + "\"" + sipMessage + "\""
//
//
//  for i := 1; i <= 10; i++ {
//      producer.Log(fmt.Sprintf("%08d", i), rawMessage)
//  }
//
//  fmt.Println("Log Over")
//
//  // os signal
//  sigterm := make(chan os.Signal, 1)
//  signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
//  <-sigterm
//}
