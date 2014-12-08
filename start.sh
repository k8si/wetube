#!/bin/bash

usage="./start.sh [your-public-ip-address]"
echo $usage

echo "starting gui..."
$WETUBE_ROOT/gui &
guiPID=$!
echo "gui pid: $guiPID"

echo "starting server..."
DEBUG=app $WETUBE_ROOT/peer/app/bin/www &
serverPID=$!
echo "server pid: $serverPID"

echo "starting wetube peer..."
MYIP=$1
echo $MYIP
$WETUBE_ROOT/wetube --ip=$MYIP &
wtPID=$!
echo "wetube client pid: $wtPID"

echo "WETUBE started."
echo "open browser to http://localhost:8080"
wait
echo "exited."
