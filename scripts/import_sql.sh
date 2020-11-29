#!/bin/bash

host="127.0.0.1"
port="3306"
user="root"
password="123456"
db_name="centnet_cdrrs"

sql_create_db="CREATE DATABASE IF NOT EXISTS ${db_name} DEFAULT CHARSET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;"
mysql  -h${host} -P${port} -u${user} -p${password} -e "${sql_create_db}"
mysql  -h${host} -P${port} -u${user} -p${password} ${db_name} < ./phone_position.sql
mysql  -h${host} -P${port} -u${user} -p${password} ${db_name} < ./voip_restored_cdr.sql
