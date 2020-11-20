package kafka

import "testing"

func TestAnalyticConsumer_Run(t *testing.T) {
	consumer := NewConsumer(&ConsumerConfig{
		Topic:       "SipPacket",
		splitTopic:  nil,
		Broker:      "192.168.1.205:9092",
		splitBroker: nil,
		Partition:   4,
		Replication: 1,
		Group:       "AnalyticCluster",
		Version:     "2.0.0",
		NumRoutine:  4,
	}, AnalyzePacket)

	consumer.Run()

	select {}
}
