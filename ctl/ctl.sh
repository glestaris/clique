#!/bin/bash
check() {
  echo "Checking..."

  [ ! -f ./clique-agent.pid ] && exit 1
  pid=$(cat ./clique-agent.pid)

  kill -0 $pid 2> /dev/null
  if [ $? -ne 0 ]; then
    echo "Process is not running"
    exit 1
  fi

  echo "Process is running"
  exit 0
}

start() {
  echo "Starting..."

  [ ! -f ./config.json ] && echo "Configuration file is not provided" && exit 1
  [ ! -f ./clique-agent ] && echo "Clique agent binary is not found" && exit 1

  $BASH_SOURCE check > /dev/null 2> /dev/null
  [ $? -eq 0 ] && echo "Agent is running already" && exit 1

  LD_LIBRARY_PATH=$PWD:$LD_LIBRARY_PATH \
    ./clique-agent -config=./config.json \
    > ./stdout.log 2> ./stderr.log < /dev/null &
  echo $! > ./clique-agent.pid

  echo "Process is started"
  exit 0
}

stop() {
  echo "Stopping..."

  $BASH_SOURCE check > /dev/null 2> /dev/null
  [ $? -ne 0 ] && echo "Agent is not running" && exit 1
  pid=$(cat ./clique-agent.pid)

  kill -s term $pid
  while kill -0 $pid 2> /dev/null; do
    sleep 0.5
  done

  echo "Process is terminated"
  exit 0
}

command=${1:-"check"}
case $command in
"check")
  check
  ;;

"start")
  start
  ;;

"stop")
  stop
  ;;

*)
  echo "Unknown command $command"
  exit 1
  ;;
esac
