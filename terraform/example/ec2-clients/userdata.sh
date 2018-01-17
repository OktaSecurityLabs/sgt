#!/bin/bash  -x
exec > ~/log.log 2>&1
sleep 60


apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 1484120AC4E9F8A1A577AEEE97A80C63C9D8B80B
add-apt-repository "deb [arch=amd64] https://osquery-packages.s3.amazonaws.com/deb deb main"
apt-get update
apt-get install osquery -y

mv /home/ubuntu/osquery.secret /etc/osquery/osquery.secret
mv /home/ubuntu/cert_bundle.pem /etc/osquery/cert_bundle.pem
cp /home/ubuntu/osquery.flags.default /etc/osquery/osquery.flags
mv /home/ubuntu/osquery.flags.default /etc/osquery/osquery.flags.default

service osqueryd start