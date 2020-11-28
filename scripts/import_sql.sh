#!/bin/bash

host="127.0.0.1"
port="3306"
user="root"
password="root"
db_name="centnet_cdrrs"

sql_create_db="CREATE DATABASE IF NOT EXISTS ${db_name} DEFAULT CHARSET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;"
mysql  -h${host} -P${port} -u${user} -p${password} < -e "${sql_create_db}"
mysql  -h${host} -P${port} -u${user} -p${password} < ./phone_position.sql
mysql  -h${host} -P${port} -u${user} -p${password} < ./voip_restored_cdr.sql
