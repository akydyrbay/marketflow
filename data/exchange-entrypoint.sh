#!/bin/bash

PORT=$1
EXCHANGE_NAME=$2

echo "Starting exchange server $EXCHANGE_NAME on port $PORT..."

while true; do
  PRICE=$(awk -v min=100 -v max=1000 'BEGIN{srand(); print min+rand()*(max-min)}')
  PAIR="BTCUSDT"
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  JSON="{\"exchange\":\"$EXCHANGE_NAME\",\"pair\":\"$PAIR\",\"price\":$PRICE,\"time\":\"$TIMESTAMP\"}"

  echo "$JSON"
  echo "$JSON" | nc -lk -p $PORT
  sleep 1
done
