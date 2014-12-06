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
openssl req -new -x509 -key $dir/id_rsa -out $dir/cacert.pem -days 1059 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$email"
