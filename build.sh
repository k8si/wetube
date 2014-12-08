#!/bin/bash

echo "REQUIREMENTS: go, node, npm, openssl"
usage="./build.sh [your-email] [this-computer's.hostname]"
echo "this-computer's.hostname: run 'host [your.public.ip.addr]'"
export WETUBE_ROOT=$PWD
export GOPATH=$WETUBE_ROOT/peer
echo "go building..."
go get golang.org/x/net/websocket
cd $GOPATH/src/helper
go build
go install
cd -
cd $GOPATH/src/peer
go build -o $WETUBE_ROOT/wetube -v director.go incoming.go outgoing.go objects.go peer.go voting.go
cd -
cd $GOPATH/src/gui
go build -o $WETUBE_ROOT/gui -v handlegui.go
cd -
echo "building server app..."
cd $GOPATH/app && npm install
cd -
echo "generating certs..."
$WETUBE_ROOT/scripts/makecert.sh $1 $2
echo "done!"
