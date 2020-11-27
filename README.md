### SIP部分包过滤规则

* INVITE包过滤规则

  ```
  1、UDP
  2、UDP数据以"INVITE sip:"开头
  ```

* INVITE的200 OK包过滤规则
  ```
  1、UDP
  2、UDP数据以"SIP/2.0 200 OK"开头
  3、UDP数据包含"\r\nCSeq: [0-9]* INVITE\r\n"（正则表达式）
  ```

* BYE的200 OK包过滤规则
  ```
  1、UDP
  2、UDP数据以"SIP/2.0 200 OK"开头
  3、UDP数据包含"\r\nCSeq: [0-9]* BYE\r\n"（正则表达式）
  ```

### 使用tcpreplay
```shell
    # -i: 指定网络接口
    # -l: 回放文件多少次
    # -p: 回放包速率，单位：pps
    # -M: 回放包速率，单位：Mbps
    tcpreplay -i en0 -l 1 -p 100 20201015.pcapng
```

### 带实现部分
* 数据库分表
* kafka优化
* sip消息字段优化

### kafka使用
[Kafka入门教程 Golang实现Kafka消息发送、接收](https://blog.csdn.net/tflasd1157/article/details/81985722)
**一、Mac版安装**
```brew install kafka```
安装kafka是需要依赖于zookeeper的，所以安装kafka的时候也会包含zookeeper
kafka的安装目录：/usr/local/Cellar/kafka
kafka的配置文件目录：/usr/local/etc/kafka
kafka服务的配置文件：/usr/local/etc/kafka/server.properties
zookeeper配置文件： /usr/local/etc/kafka/zookeeper.properties

```
# server.properties中的重要配置

broker.id=0
listeners=PLAINTEXT://:9092
advertised.listeners=PLAINTEXT://127.0.0.1:9092
log.dirs=/usr/local/var/lib/kafka-logs
```
```
# zookeeper.properties

dataDir=/usr/local/var/lib/zookeeper
clientPort=2181
maxClientCnxns=0
```

二： 启动zookeeper
```
# 新起一个终端启动zookeeper
cd /usr/local/Cellar/kafka/1.0.0
./bin/zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties
```

三： 启动kafka
```
# 新起一个终端启动zookeeper，注意启动kafka之前先启动zookeeper
cd /usr/local/Cellar/kafka/1.0.0
./bin/kafka-server-start /usr/local/etc/kafka/server.properties
```

四：创建topic
```
# 新起一个终端来创建主题
cd /usr/local/Cellar/kafka/1.0.0

# 创建一个名为“test”的主题，该主题有1个分区
./bin/kafka-topics --create
    --zookeeper localhost:2181
    --partitions 1
    --topic test
```

五：查看topic
```
# 创建成功可以通过 list 列举所有的主题
./bin/kafka-topics --list --zookeeper localhost:2181

# 查看某个主题的信息
./bin/kafka-topics --describe --zookeeper localhost:2181 --topic <name>
```

六：发送消息
```
# 新起一个终端，作为生产者，用于发送消息，每一行算一条消息，将消息发送到kafka服务器
  > ./bin/kafka-console-producer --broker-list localhost:9092 --topic test
  This is a message
  This is another message
```

七：消费消息(接收消息)
```
# 新起一个终端作为消费者，接收消息
cd /usr/local/Cellar/kafka/1.0.0
> ./bin/kafka-console-consumer --bootstrap-server localhost:9092 --topic test --from-beginning
This is a message
This is another message
```