package kafka

// 参考: https://www.cnblogs.com/mrblue/p/10770651.html

import (
	"centnet-cdrrs/library/log"
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"strings"
	"time"
)

type ConsumerHandler func(*Consumer, interface{})

// Consumer Consumer配置
type ConsumerConfig struct {
	Topic       string
	splitTopic  []string
	Broker      string
	splitBroker []string
	Partition   int32
	Replication int16
	Group       string
	Version     string

	NumRoutine int
}

type Consumer struct {
	sc   *sarama.Config
	cc   *ConsumerConfig
	fun  ConsumerHandler
	next *Producer
}

func NewConsumer(cc *ConsumerConfig, fun ConsumerHandler) *Consumer {
	return newConsumer(cc, fun)
}

func newConsumer(cc *ConsumerConfig, fun ConsumerHandler) *Consumer {
	ver, err := sarama.ParseKafkaVersion(cc.Version)
	if err != nil {
		log.Error(err)
		return nil
	}

	if cc.splitTopic = strings.Split(cc.Topic, ","); len(cc.splitTopic) == 0 {
		log.Errorf("invalid arguments: WrappedConsumer.cc.Topic = %s\n", cc.Topic)
		return nil
	}
	if cc.splitBroker = strings.Split(cc.Broker, ","); len(cc.splitBroker) == 0 {
		log.Errorf("invalid arguments: WrappedConsumer.cc.Broker = %s\n", cc.Broker)
		return nil
	}

	sc := sarama.NewConfig()
	sc.Version = ver

	return &Consumer{
		sc:  sc,
		cc:  cc,
		fun: fun,
	}
}

func (ac *Consumer) SetNextProducer(producer *Producer) {
	ac.next = producer
}

func (ac *Consumer) Run() error {
	numRoutine := ac.cc.NumRoutine
	//ctx, cancel := make([]context.Context, numRoutine), make([]context.CancelFunc, numRoutine)
	ctx := make([]context.Context, numRoutine)

	for i := 0; i < numRoutine; i++ {
		ctx[i], _ = context.WithCancel(context.Background())
	}

	cgHandler := consumerGroupHandler{fun: ac.fun, customConsumer: ac}
	group, err := sarama.NewConsumerGroup(ac.cc.splitBroker, ac.cc.Group, ac.sc)
	if err != nil {
		panic(err)
	}
	//defer func() { _ = group.Close() }()

	for i := 0; i < numRoutine; i++ {
		go func(i int) {
			for {
				err := group.Consume(ctx[i], ac.cc.splitTopic, &cgHandler)
				if err != nil {
					log.Error(err)
					time.Sleep(time.Second * 5)
				}
			}
		}(i)
	}

	return nil
}

func (ac *Consumer) DeleteTopic(topic string) {
	ver, err := sarama.ParseKafkaVersion(ac.cc.Version)
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Version = ver

	admin, err := sarama.NewClusterAdmin(strings.Split(ac.cc.Broker, ","), config)
	if err != nil {
		panic(err)
	}

	if err = admin.DeleteTopic(topic); err != nil {
		panic(err)
	}
	log.Debug("topic '%s' deleted", topic)

	if err := admin.Close(); err != nil {
		panic(err)
	}
}

func (ac *Consumer) CreateTopic() {
	ver, err := sarama.ParseKafkaVersion(ac.cc.Version)
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Version = ver

	admin, err := sarama.NewClusterAdmin(strings.Split(ac.cc.Broker, ","), config)
	if err != nil {
		panic(err)
	}

	detail, err := admin.ListTopics()
	if err != nil {
		panic(err)
	}

	for _, v := range ac.cc.splitTopic {
		if d, ok := detail[v]; ok {
			if ac.cc.Partition > d.NumPartitions {
				if err := admin.CreatePartitions(v, ac.cc.Partition, nil, false); err != nil {
					panic(err)
				}
				log.Debug("topic '%s' partition '%d' / NumPartitions '%d' created",
					v, ac.cc.Partition, d.NumPartitions)
			}
		} else {
			if err := admin.CreateTopic(v, &sarama.TopicDetail{
				NumPartitions:     ac.cc.Partition,
				ReplicationFactor: ac.cc.Replication,
			}, false); err != nil {
				panic(err)
			}
			log.Debug("topic '%s' created", v)
		}
	}

	if detail, err := admin.ListTopics(); err != nil {
		fmt.Println(err)
	} else {
		for k := range detail {
			log.Debug("[%s] %+v", k, detail[k])
		}
	}

	if err := admin.Close(); err != nil {
		panic(err)
	}
}

type consumerGroupHandler struct {
	fun            ConsumerHandler
	customConsumer *Consumer
}

func (c *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	count := 0
	picker := time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case t := <-picker.C:
				log.Debug(t.Format("2006-01-02 15:04:05.000000"), claim.Partition(), count)
			}
		}
	}()

	for msg := range claim.Messages() {
		count++

		c.fun(c.customConsumer, msg.Value)

		//key, value := string(msg.Key), string(msg.Value)
		//fmt.Println(key, value, msg.Partition, count)
		session.MarkMessage(msg, "")
	}

	return nil
}
