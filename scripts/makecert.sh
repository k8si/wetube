#!/bin/bash
# stolen from: https://gist.github.com/spikebike/2232102
# usage: ./makecert.sh [email]
usage="./makecert.sh [email] [this-computer's-hostname]"
echo $usage
echo "(note that this requires openssl)"
echo $WETUBE_ROOT
# export WETUBE_ROOT=$PWD
# dir=$WETUBE_ROOT/peer/src/peer
dir=$WETUBE_ROOT
email=$1
hostname=$2
echo "generating rsa keys..."
ssh-keygen -t rsa -C $email -f $dir/id_rsa
echo "generating x509 cert req..."
hostname=$2
openssl req -new -x509 -key $dir/id_rsa -out $dir/server_cert.pem -days 1059 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=$hostname/emailAddress=$email"
# copy/rename to maximize the number of files on your computer
cp $dir/id_rsa $dir/server_key.pem
# convert privkey to DER format because *reasons*
openssl rsa -in $dir/server_key.pem -outform DER -out $dir/keyout.der

