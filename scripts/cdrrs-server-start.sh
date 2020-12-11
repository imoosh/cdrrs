#!/bin/bash

ulimit -n 65535

SERVICE=./bin/centnet-cdrrs
CONFIG="./conf/config.toml"

PID=$(ps -ef | grep ${SERVICE} | grep -v grep | tr -s ' ' | cut -d ' ' -f 2)
if [[ -n "${PID}" ]]; then
  echo "${SERVICE} is running."
  exit
fi

case $1 in
-daemon)
  nohup ${SERVICE} -c ${CONFIG} 1> /dev/null 2>&1 &
  ;;
*)
  ${SERVICE} -c ${CONFIG}
  ;;
esac
