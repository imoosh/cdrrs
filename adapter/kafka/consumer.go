package kafka

import (
	"centnet-cdrrs/library/log"
	"context"
	"github.com/Shopify/sarama"
	"strings"
	"time"
)

// https://www.cnblogs.com/mrblue/p/10770651.html
// https://www.yuque.com/sanweishe/pqy91r/osgq9q
// github.com/!shopify/sarama@v1.27.2/functional_consumer_group_test.go

// Consumer Consumer配置
type ConsumerConfig struct {
	Topic  string
	Broker string
	Group  string

	GroupMembers        int
	FlowRateFlushPeriod int
}

type ConsumerHandler func(*ConsumerGroupMember, interface{}, interface{})

type ConsumerGroupMember struct {
	sarama.ConsumerGroup
	ClientID string
	errs     []error

	Next       *Producer
	handle     ConsumerHandler
	TotalCount uint64
	LastCount  uint64
	Timer      *time.Timer
	Conf       *ConsumerConfig
}

func NewConsumerGroupMember(c *ConsumerConfig, clientID string, handle ConsumerHandler) *ConsumerGroupMember {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.ClientID = clientID
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Timeout = 10 * time.Second

	group, err := sarama.NewConsumerGroup(strings.Split(c.Broker, ","), c.Group, config)
	if err != nil {
		log.Error(err)
		return nil
	}

	member := &ConsumerGroupMember{
		ConsumerGroup: group,
		ClientID:      clientID,
		handle:        handle,
		TotalCount:    0,
		LastCount:     0,
		Timer:         time.NewTimer(time.Second * time.Duration(c.FlowRateFlushPeriod)),
		Conf:          c,
	}
	go member.loop(strings.Split(c.Topic, ","))
	return member
}

func (m *ConsumerGroupMember) loop(topics []string) {
	// 处理错误
	go func() {
		for err := range m.Errors() {
			_ = m.Close()
			m.errs = append(m.errs, err)
		}
	}()

	// 统计流量
	go func() {
		for {
			select {
			case <-m.Timer.C:
				log.Debugf("%s flow rate: %d pps, total: %d", m.ClientID, (m.TotalCount-m.LastCount)/uint64(m.Conf.FlowRateFlushPeriod), m.TotalCount)
				m.LastCount = m.TotalCount
				m.Timer.Reset(time.Second * time.Duration(m.Conf.FlowRateFlushPeriod))
			}
		}
	}()

	// 循环消费
	ctx := context.Background()
	for {
		if err := m.Consume(ctx, topics, m); err == sarama.ErrClosedConsumerGroup {
			return
		} else if err != nil {
			m.errs = append(m.errs, err)
			return
		}
	}
}

func (m *ConsumerGroupMember) Stop() { _ = m.Close() }

func (m *ConsumerGroupMember) Setup(s sarama.ConsumerGroupSession) error {
	// enter post-setup state
	return nil
}

func (m *ConsumerGroupMember) Cleanup(s sarama.ConsumerGroupSession) error {
	// enter post-cleanup state
	return nil
}

func (m *ConsumerGroupMember) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {

	count := make(map[int]int)
	for msg := range c.Messages() {
		// TODO
		m.handle(m, msg.Key, msg.Value)
		count[int(msg.Partition)]++
		//log.Debugf("ClientID: %s, Topic: %s, Partition: %d, Offset: %d, TotalCount: %d", m.ClientID, msg.Topic, msg.Partition, msg.Offset, count[int(msg.Partition)])

		s.MarkMessage(msg, "")
	}
	return nil
}

// 设置下一级流水线作业
func (m *ConsumerGroupMember) SetNextPipeline(producer *Producer) {
	m.Next = producer
}
