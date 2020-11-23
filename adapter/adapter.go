package adapter

//const (
//	PacketSnifferAdapter = iota
//	FormattedFileAdapter
//	KafkaConsumerAdapter
//)
//
//type DataCollectionAdapterType int32
//type DataCollectionAdapterCustomFunc func(interface{})
//
//type Consumer interface {
//	Run() error
//}
//
//func NewConsumer(t DataCollectionAdapterType, params interface{}) Consumer {
//	switch t {
//	case PacketSnifferAdapter:
//		return sniffer.NewPacketSniffer(params.(*sniffer.Conf))
//	case FormattedFileAdapter:
//
//	case KafkaConsumerAdapter:
//		return analytic.NewKafkaConsumer(params.(*analytic.ConsumerConfig))
//	}
//	return nil
//}
