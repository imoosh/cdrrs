最新版本CentOS默认安装的是MariaDB，这个是MySQL分支。为了需要，还是要在系统中安装MySQL，而且安装完成之后可以覆盖MariaDB。

### 下载安装MySQL官方yum repository

```
# 下载yum repository
wget -i -c http://dev.mysql.com/get/mysql57-community-release-el7-10.noarch.rpm

# 安装yum repository
yum -y install mysql57-community-release-el7-10.noarch.rpm

# 安装mysql服务
yum -y install mysql-community-server
```

### MySQL数据库配置

- 启动MySQL

```shell
systemctl start  mysqld.service 
```

- 查看运行状态

```shell
systemctl status mysqld.service
```

- 获取默认密码

MySQL安装过程中，系统会为root用户设置一个默认登陆密码，通过以下命令在日志文件中获取

```shell
grep "password" /var/log/mysqld.log
```

- 使用root默认密码进入数据库

```shell
# 使用上一步获取的密码替换PASSWORD
mysql -uroot -p'PASSWORD'
```

- 修改默认密码

```sql
ALTER USER 'root'@'localhost' IDENTIFIED BY 'new password';
```

'new password'替换为要设置的新密码，这里的密码设置必须包含：大小写字母、数字和特殊字符(,/@;:#等)，不然会报错。

- 设置弱密码

如果要设置弱密码，需要做以下配置，先查看密码策略，可以看到`validate_password_policy`为`MEDIUM`。

```sql
mysql> show variables like '%password%';
+----------------------------------------+-----------------+
| Variable_name                          | Value           |
+----------------------------------------+-----------------+
| default_password_lifetime              | 0               |
| disconnect_on_expired_password         | ON              |
| log_builtin_as_identified_by_password  | OFF             |
| mysql_native_password_proxy_users      | OFF             |
| old_passwords                          | 0               |
| report_password                        |                 |
| sha256_password_auto_generate_rsa_keys | ON              |
| sha256_password_private_key_path       | private_key.pem |
| sha256_password_proxy_users            | OFF             |
| sha256_password_public_key_path        | public_key.pem  |
| validate_password_check_user_name      | OFF             |
| validate_password_dictionary_file      |                 |
| validate_password_length               | 8               |
| validate_password_mixed_case_count     | 1               |
| validate_password_number_count         | 1               |
| validate_password_policy               | MEDIUM          |
| validate_password_special_char_count   | 1               |
+----------------------------------------+-----------------+
```

打开`/etc/my.cnf`文件修改密码策略，添加`validate_password_policy`配置，选择0(LOW)，1(MEDIUM)，2(STRONG)其中一种，选择2需要提供密码字典文件。

```shell
#添加validate_password_policy配置
validate_password_policy=0

#关闭密码策略
validate_password=off
```

- 重启MySQL服务

```
systemctl restart mysqld
```



### 开启MySQL远程访问

执行以下命令开启远程访问权限，命令中开启的IP是`192.168.0.1`，如果开启所有的，用%代替IP即可。

```sql
grant all privileges on *.* to 'root'@'192.168.0.1' identified by 'password' with grant option;

# 刷新
flush privileges; 
```



### 修改MySQL字符编码

显示原先编码

```sql
mysql> show variables like '%character%';
+--------------------------+----------------------------+
| Variable_name            | Value                      |
+--------------------------+----------------------------+
| character_set_client     | utf8                       |
| character_set_connection | utf8                       |
| character_set_database   | latin1                     |
| character_set_filesystem | binary                     |
| character_set_results    | utf8                       |
| character_set_server     | latin1                     |
| character_set_system     | utf8                       |
| character_sets_dir       | /usr/share/mysql/charsets/ |
+--------------------------+----------------------------+
8 rows in set (0.00 sec)
```

修改/etc/my.cnf

```
[mysqld]
character_set_server=utf8
init_connect='SET NAMES utf8'
```

确认修改成功

```sql
mysql> show variables like '%character%';
+--------------------------+----------------------------+
| Variable_name            | Value                      |
+--------------------------+----------------------------+
| character_set_client     | utf8                       |
| character_set_connection | utf8                       |
| character_set_database   | utf8                       |
| character_set_filesystem | binary                     |
| character_set_results    | utf8                       |
| character_set_server     | utf8                       |
| character_set_system     | utf8                       |
| character_sets_dir       | /usr/share/mysql/charsets/ |
+--------------------------+----------------------------+
8 rows in set (0.00 sec)
```



### 设置firewalld开放端口

添加MySQL端口3306

```shell
firewall-cmd --zone=public --add-port=3306/tcp --permanent

# 从新载入
firewall-cmd --reload
```



### MySQL修改默认端口

查看端口号

```sql
mysql> show global variables like 'port';
+---------------+-------+
| Variable_name | Value |
+---------------+-------+
| port          | 3306  |
+---------------+-------+
1 row in set (0.01 sec)
```

修改端口号

打开/etc/my.cnf文件，增加端口参数，设置端口

```shell
[mysqld]
port=33066
#
# Remove leading # and set to the amount of RAM for the most important data
# cache in MySQL. Start at 70% of total RAM for dedicated server, else 10%.
# innodb_buffer_pool_size = 128M
#
# Remove leading # to turn on a very important data integrity option: logging
# changes to the binary log between backups.
# log_bin
#
# Remove leading # to set options mainly useful for reporting servers.
# The server defaults are faster for transactions and fast SELECTs.
# Adjust sizes as needed, experiment to find the optimal values.
# join_buffer_size = 128M
# sort_buffer_size = 2M
# read_rnd_buffer_size = 2M
datadir=/var/lib/mysql
socket=/var/lib/mysql/mysql.sock

# Disabling symbolic-links is recommended to prevent assorted security risks
symbolic-links=0

log-error=/var/log/mysqld.log
pid-file=/var/run/mysqld/mysqld.pid

character_set_server=utf8
init_connect='SET NAMES utf8'
```

修改完重启MySQL服务

```shell
systemctl restart mysqld.service
```

