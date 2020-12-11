#!/bin/bash

USER='centnet_cdrrs'
PASSWORD='123456'
DATABASE='centnet_cdrrs'

CREATE_USER="create user 'centnet_cdrrs'@'%' identified by '123456';"+\
            "create user 'centnet_cdrrs'@'localhost' identified by '123456';"+\
            "grant all on *.* to 'centnet_cdrrs'@'localhost';"+\
            "grant all on *.* to 'centnet_cdrrs'@'%';"+\
            "flush privileges;"

CREATE_DB="CREATE DATABASE IF NOT EXISTS ${DATABASE}"+\
          "DEFAULT CHARACTER SET utf8mb4"+\
          "DEFAULT COLLATE utf8mb4_general_ci;"

mysql -uroot -p mysql -e "${CREATE_USER}"
mysql -uroot -p -e "${CREATE_DB}"