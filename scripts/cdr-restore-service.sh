#!/bin/bash

ulimit -n 65535

if [ $# != 1 ]; then
  echo "usage: ./cdr-restore-service.sh {start|stop|version}"
  exit 1
fi

SERVICE=./bin/cdr-restore-service
CONFIG_FILE="./conf/config.toml"

# Mac OS X 操作系统
if [ "$(uname)" == "Darwin" ]; then
  BIN="./bin/vs_darwin_amd64"
# GNU/Linux操作系统
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
  BIN="./bin/vs_linux_amd64"
# Windows NT操作系统
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
  BIN=""
fi

# shellcheck disable=SC2006
# shellcheck disable=SC2009
IsExist=$(ps -ef | grep ${SERVICE} | grep -v grep | tr -s ' ')
if [ -n "${IsExist}" ]; then
  echo "${SERVICE} is running."
  exit
fi

case $1 in
start)
  ${SERVICE} -c $CONFIG_FILE
  ;;
stop)
  if [ -n "${IsExist}" ]; then
    pkill -9 "${SERVICE}"
  fi
  ;;
v | ver | version)
  ${SERVICE} -v
  ;;
*) ;;

esac
