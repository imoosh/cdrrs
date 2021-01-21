### 安装kafka

#### 更新系统

```shell
sudo yum install epel-release -y
sudo yum update -y　
```

#### 安装JDK

方法一：[oracle官网下载jdk](https://www.oracle.com/java/technologies/javase/javase-jdk8-downloads.html)

```shell
 sudo rpm -ivh jdk-8u241-linux-x64.rpm
```

方法二：安装openjdk

```shell
sudo yum install -y java-1.8.0-openjdk
```

检查是否安装成功

```shell
java -version
```

#### 配置Java环境配置

```shell
echo $JAVA_HOME

echo "JAVA_HOME=$(readlink -f /usr/bin/java | sed "s:bin/java::")" | sudo tee -a /etc/profile
source /etc/profile
echo $JAVA_HOME
```

#### 安装kafka

下载kafka安装包，并解压

```shell
wget https://mirrors.tuna.tsinghua.edu.cn/apache/kafka/2.4.0/kafka_2.12-2.4.0.tgz

sudo tar -zvxf kafka_2.12-2.4.0.tgz -C  /opt/

sudo ln -s /opt/kafka_2.12-2.4.0 /opt/kafka
```

#### 启动服务

```shell
cd /opt/kafka/bin

# 启动zookeeper
sudo ./zookeeper-server-start.sh ../config/zookeeper.properties

# 启动kafka
sudo ./kafka-server-start.sh ../config/server.properties
```

#### 配置系统服务单元

* zookeeper服务

```shell
vi /etc/systemd/system/zookeeper.service
[Unit]
Description=Apache Zookeeper server
Documentation=http://zookeeper.apache.org
Requires=network.target remote-fs.target
After=network.target remote-fs.target
 
[Service]
Type=simple
ExecStart=/opt/kafka/bin/zookeeper-server-start.sh /opt/kafka/config/zookeeper.properties
ExecStop=/opt/kafka/bin/zookeeper-server-stop.sh
Restart=on-abnormal
User=root
Group=root
 
[Install]
WantedBy=multi-user.target
```

验证是否生效

```shell
sudo systemctl start zookeeper
sudo systemctl status zookeeper
sudo systemctl start zookeeper
```

* kafka服务

```shell
sudo vi /etc/systemd/system/kafka.service
 
[Unit]
Description=Apache Kafka Server
Documentation=http://kafka.apache.org/documentation.html
Requires=zookeeper.service
 
[Service]
Type=simple
ExecStart=/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties
ExecStop=/opt/kafka/bin/kafka-server-stop.sh
Restart=on-abnormal
 
 
[Install]
WantedBy=multi-user.target
```

验证是否生效

```shell
sudo systemctl start kafka
sudo systemctl status kafka
sudo systemctl stop kafka
```

### 使用kafka

#### 创建topic

```shell
# 创建一个kafka topic
# replication-factor：副本数量
# partitions：分区数量
bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic testTopic

# 查看创建的topic
bin/kafka-topics.sh --list --zookeeper localhost:2181
```

#### 删除topic

```shell
# 删除topic
bin/kafka-topics.sh --delete --topic testTopic --zookeeper localhost:2181
```

#### 发送消息

```shell
# 使用生产者发送消息
bin/kafka-console-producer.sh --broker-list localhost:9092 --topic testTopic
```

#### 接受消息

```shell
# 使用消费者接受消息
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic testTopic --from-beginning
```

#### 清空topic内容

### 参考资料

* [centos 安装kafka](https://www.cnblogs.com/Hackerman/p/12595246.html)

