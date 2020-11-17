package kafka

import (
    "fmt"
    "github.com/Shopify/sarama"
    "time"
)

func Consumer(id int) {
    config := sarama.NewConfig()
    config.Consumer.Return.Errors = true
    config.Version = sarama.V2_2_0_0

    consumer, err := sarama.NewConsumer([]string{"192.168.1.205:9092"}, config)
    if err != nil {
        fmt.Printf("NewConsumer error: %s\n", err.Error())
        return
    }
    defer consumer.Close()

    partitionConsumer, err := consumer.ConsumePartition("testTopic", 0, sarama.OffsetOldest)
    if err != nil {
        fmt.Printf("ConsumePartition error: %s\n", err.Error())
        return
    }
    defer partitionConsumer.Close()

    for {
        select {
        case msg := <-partitionConsumer.Messages():
            msg.Timestamp.Add(time.Second)
            fmt.Printf("id: %d, msg offset: %d, partition: %d, timestamp: %s, value: %s\n", id,
             msg.Offset, msg.Partition, msg.Timestamp.String(), string(msg.Value))
        case err := <-partitionConsumer.Errors():
            fmt.Printf("error: %s\n", err.Error())
        }
    }
}
