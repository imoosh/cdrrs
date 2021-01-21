### 下载redis

```shell
mkdir /home/redis/
cd /home/redis/
wget http://download.redis.io/releases/redis-5.0.5.tar.gz
tar -xzvf redis-5.0.5.tar.gz
```

### 安装编译环境

```shell
yum -y install make automake cmake gcc g++
cd /home/redis/redis-5.0.5/
```

### 编译redis

```shell
make && make install
```

###  修改配置

将`/home/redis/redis-5.0.5/redis.conf`拷贝至`/usr/local/etc/`目录下

```shell
cp `/home/redis/redis-5.0.5/redis.conf` /usr/local/etc
```

修改配置如下

```shell
vim /etc/redis.conf

bind 0.0.0.0      # 绑定运行IP
protected-mode no # 关闭保护模式
daemonize yes     # 守护进程

# 注释掉持久化策略，防止redis写硬盘拖慢系统性能
#save 900 1
#save 300 10
#save 60 10000
```

### 启动服务

```shell
/usr/local/bin/redis-server /usr/local/etc/redis.conf
```

### 检查redis进程和端口

```shell
ps -ef | grep redis
netstat -tunlep | grep 6379
```