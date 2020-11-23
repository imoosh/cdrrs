package kafka

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

// Conf 配置
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

	log.Debug("Broker: ", cfg.Broker, ", Topic: ", cfg.Topic)
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
					log.Errorf("[producer] err=[%s] topic=[%s] key=[%s] val=[%s]",
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
		log.Debug("[producer]", "KEY: ", key, ", VALUE: ...")
		return
	default:
		log.Debug("[producer] err=[msgQ is full] key=[%s] val=[%s]", msg.Key, msg.Value)
	}
}
