package main

// 参考: https://www.cnblogs.com/mrblue/p/10770651.html

import (
    "context"
    "fmt"
    "github.com/Shopify/sarama"
    "os"
    "os/signal"
    "runtime"
    "strings"
    "syscall"
    "time"
)

var cc = ConsumerConfig{
    Topic:       []string{"SipPacket"},
    Broker:      "192.168.1.205:9092",
    Partition:   6,
    Replication: 1,
    Group:       "AnalyticCluster",
    Version:     "2.2.0",
}

// Consumer Consumer配置
type ConsumerConfig struct {
    Topic       []string `xml:"topic"`
    Broker      string   `xml:"broker"`
    Partition   int32    `xml:"partition"`
    Replication int16    `xml:"replication"`
    Group       string   `xml:"group"`
    Version     string   `xml:"version"`
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    ver, err := sarama.ParseKafkaVersion(cc.Version)
    if err != nil {
        panic(err)
    }

    config := sarama.NewConfig()
    config.Version = ver

    //DeleteTopic(config, "SipPacket")
    CreateTopic(config)

    const ChannelSize = 6
    var ctx [ChannelSize]context.Context
    var cancel [ChannelSize]context.CancelFunc
    for i := 0; i < ChannelSize; i++ {
        ctx[i], cancel[i] = context.WithCancel(context.Background())
    }
    //ctx, cancel := context.WithCancel(context.Background())

    consumer := Consumer{}
    group, err := sarama.NewConsumerGroup(strings.Split(cc.Broker, ","), cc.Group, config)
    if err != nil {
        panic(err)
    }
    defer func() { _ = group.Close() }()

    for i := 0; i < ChannelSize; i++ {
        go func(i int) {
            for {
                err := group.Consume(ctx[i], cc.Topic, &consumer)
                if err != nil {
                    fmt.Println(err)
                    time.Sleep(time.Second * 5)
                }
            }
        }(i)
    }

    // os signal
    sigterm := make(chan os.Signal, 1)
    signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
    <-sigterm

    for i := range cancel {
        cancel[i]()
    }
}

func DeleteTopic(config *sarama.Config, topic string) {
    admin, err := sarama.NewClusterAdmin(strings.Split(cc.Broker, ","), config)
    if err != nil {
        panic(err)
    }

    if err = admin.DeleteTopic(topic); err != nil {
        panic(err)
    }
    fmt.Printf("topic '%s' deleted\n", topic)

    if err := admin.Close(); err != nil {
        panic(err)
    }
}

func CreateTopic(config *sarama.Config) {
    admin, err := sarama.NewClusterAdmin(strings.Split(cc.Broker, ","), config)
    if err != nil {
        panic(err)
    }

    detail, err := admin.ListTopics()
    if err != nil {
        panic(err)
    }

    for _, v := range cc.Topic {
        if d, ok := detail[v]; ok {
            if cc.Partition > d.NumPartitions {
                if err := admin.CreatePartitions(v, cc.Partition, nil, false); err != nil {
                    panic(err)
                }
                fmt.Printf("topic '%s' partition '%d' / NumPartitions '%d' created\n",
                    v, cc.Partition, d.NumPartitions)
            }
        } else {
            if err := admin.CreateTopic(v, &sarama.TopicDetail{
                NumPartitions:     cc.Partition,
                ReplicationFactor: cc.Replication,
            }, false); err != nil {
                panic(err)
            }
            fmt.Printf("topic '%s' created\n", v)
        }
    }

    if detail, err := admin.ListTopics(); err != nil {
        fmt.Println(err)
    } else {
        for k := range detail {
            fmt.Printf("[%s] %+v\n", k, detail[k])
        }
    }

    if err := admin.Close(); err != nil {
        panic(err)
    }
}

type Consumer struct {
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
    return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
    return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    count := 0
    picker := time.NewTicker(time.Second)

    go func() {
        for {
            select {
            case t := <-picker.C:
                fmt.Println(t.Format("2006-01-02 15:04:05.000000"), claim.Partition(), count)
            }
        }
    }()

    for msg := range claim.Messages() {
        count++
        //key, value := string(msg.Key), string(msg.Value)
        //fmt.Println(key, value, msg.Partition, count)
        session.MarkMessage(msg, "")
    }

    return nil
}
