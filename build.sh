#!/bin/bash

echo "REQUIREMENTS: go, node, npm, openssl"
usage="./build.sh [your-email] [this-computer's.hostname]"
echo "this-computer's.hostname: run 'host [your.public.ip.addr]'"
export WETUBE_ROOT=$PWD
export GOPATH=$WETUBE_ROOT/peer
cd $GOPATH/src/peer
echo "go building..."
go build -v -i -o $WETUBE_ROOT/wetube director.go incoming.go outgoing.go objects.go peer.go voting.go
cd -
cd $GOPATH/src/gui
go build -v -i -o $WETUBE_ROOT/gui handlegui.go
cd -
echo "building server app..."
cd $GOPATH/app && npm install
cd -
echo "generating certs..."
$WETUBE_ROOT/scripts/makecert.sh $1 $2
echo "done!"
