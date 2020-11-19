package adapter

import (
	"centnet-cdrrs/adapter/kafka/analytic"
	"centnet-cdrrs/adapter/sniffer"
)

const (
	PacketSnifferAdapter = iota
	FormattedFileAdapter
	KafkaConsumerAdapter
)

type DataCollectionAdapterType int32
type DataCollectionAdapterCustomFunc func(interface{})

type Adapter interface {
	Run() error
}

func NewAdapter(t DataCollectionAdapterType, params interface{}) Adapter {
	switch t {
	case PacketSnifferAdapter:
		return sniffer.NewPacketSniffer(params.(*sniffer.Config))
	case FormattedFileAdapter:

	case KafkaConsumerAdapter:
		return analytic.NewKafkaConsumer(params.(*analytic.ConsumerConfig))
	}
	return nil
}
