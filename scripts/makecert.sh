#!/bin/bash
# stolen from: https://gist.github.com/spikebike/2232102
# usage: ./makecert.sh [email]
usage="./makecert.sh [your-email] [hostname]"
echo $usage
dir=$WETUBE_ROOT/peer/src/peer
email=$1
echo "gen rsa keys"
ssh-keygen -t rsa -C $email -f $dir/id_rsa
echo "gen x509 cert"
hostname=$2
openssl req -new -x509 -key $dir/id_rsa -out $dir/server_cert.pem -days 1059 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=$hostname/emailAddress=$email"
