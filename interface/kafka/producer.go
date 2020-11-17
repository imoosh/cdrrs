package kafka

import (
    "fmt"
    "github.com/Shopify/sarama"
)

func Producer() {
    fmt.Printf("producer_test\n")
    config := sarama.NewConfig()

    //acks为0：这意味着producer发送数据后，不会等待broker确认，直接发送下一条数据，性能最快
    //acks为1：为1意味着producer发送数据后，需要等待leader副本确认接收后，才会发送下一条数据，性能中等
    //acks为-1：这个代表的是all，意味着发送的消息写入所有的ISR集合中的副本（注意不是全部副本）后，才会发送下一条数据，性能最慢，但可靠性最强
    config.Producer.RequiredAcks = sarama.NoResponse
    config.Producer.Partitioner = sarama.NewRandomPartitioner
    config.Producer.Return.Successes = true
    config.Producer.Return.Errors = true
    config.Version = sarama.V2_2_0_0

    producer, err := sarama.NewAsyncProducer([]string{"192.168.1.205:9092"}, config)
    if err != nil {
        fmt.Printf("producer_test create producer error :%s\n", err.Error())
        return
    }

    defer producer.AsyncClose()

    // send message
    msg := &sarama.ProducerMessage{
        Topic: "testTopic",
        Key:   sarama.StringEncoder("go_test"),
    }

    value := "this is message"
    for i := 0; i < 4; i++{
        //fmt.Scanln(&value)
        msg.Value = sarama.ByteEncoder(value + fmt.Sprintf(" -- %04d", i))
        msg.Partition = int32(i)
        fmt.Printf("input [%s] %s\n", value, fmt.Sprintf(" -- %04d", i))

        // send to chain
        producer.Input() <- msg

        select {
        case suc := <-producer.Successes():
           fmt.Printf("offset: %d,  timestamp: %s\n", suc.Offset, suc.Timestamp.String())
        case fail := <-producer.Errors():
           fmt.Printf("err: %s\n", fail.Err.Error())
        }
    }
}
