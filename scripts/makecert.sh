#!/bin/bash
# stolen from: https://gist.github.com/spikebike/2232102
# usage: ./makecert.sh [email]
usage="./makecert.sh [your-email]"
echo $usage
dir=$WETUBE_ROOT/peer/src/peer
email=$1
echo "gen rsa keys"
ssh-keygen -t rsa -C $email -f $dir/id_rsa
echo "gen x509 cert"
openssl req -new -x509 -key $dir/id_rsa -out $dir/cacert.pem -days 1059 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=174.62.219.8/emailAddress=$email"


# dir=$WETUBE_ROOT/peer/src/peer/certs
# echo $dir
# mkdir $dir
# echo "make home certs"
# openssl req -new -nodes -x509 -out $dir/home.pem -keyout $dir/home.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=174.62.219.8/emailAddress=kate@home"
# echo "make ec2 cert"
# openssl req -new -nodes -x509 -out $dir/ec2.pem -keyout $dir/ec2.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=128.119.243.164/emailAddress=kate@ec2"
# echo "make edlab cert"
# openssl req -new -nodes -x509 -out $dir/edlab.pem -keyout $dir/edlab.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=54.149.118.210/emailAddress=kate@edlab"