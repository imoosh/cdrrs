#!/bin/bash

SERVICE=./bin/centnet-cdrrs

PID=$(ps -ef | grep ${SERVICE} | grep -v grep | tr -s ' ' | cut -d ' ' -f 2)
if [[ -n "${PID}" ]]; then
  kill -9 "${PID}"
  echo "${SERVICE}[$PID] exit."
  exit
fi
