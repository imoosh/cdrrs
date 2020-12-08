#!/bin/bash

# 部署完kafka和zookeeper后，记得修改数据路径
# 启动zookeeper
./zookeeper-server-start.sh -daemon ../config/zookeeper.properties

# 启动kafka
./kafka-server-start.sh -daemon ../config/server.properties

KAFKA_TOP=/opt/kafka

# 查看已有的topic
${KAFKA_TOP}/bin/kafka-topics.sh --list --zookeeper localhost:2181

# 删除已有的topic
#${KAFKA_TOP}/bin/kafka-topics.sh --delete --topic RawVoipPacket --zookeeper localhost:2181
#${KAFKA_TOP}/bin/kafka-topics.sh --delete --topic SemiFinishedCDR --zookeeper localhost:2181

# 创建topic
${KAFKA_TOP}/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic RawVoipPacket
${KAFKA_TOP}/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic SemiFinishedCDR

