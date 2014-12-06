#!/bin/bash
# stolen from: https://gist.github.com/spikebike/2232102
# usage: ./makecert.sh [email]
dir=$WETUBE_ROOT/peer/src/peer/certs
echo $dir
mkdir $dir
echo "make server cert"
openssl req -new -nodes -x509 -out $dir/server.pem -keyout $dir/server.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$1"
echo "make client cert"
openssl req -new -nodes -x509 -out $dir/client.pem -keyout $dir/client.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$1"