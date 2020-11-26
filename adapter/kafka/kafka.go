package kafka

type Config struct {
	/* 原始数据包缓存配置 */
	SipPacketProducer *ProducerConfig

	/* 原始数据包读取配置 */
	SipPacketConsumer *ConsumerConfig

	/* 解析的数据包缓存配置 */
	RestoreCDRProducer *ProducerConfig

	/* 解析的数据包读取配置 */
	RestoreCDRConsumer *ConsumerConfig

	/* 话单数据推送配置 */
	FraudModelProducer *ProducerConfig
}
