#!/bin/bash -x
exec > ~/log.log 2>&1

apt update
#apt upgrade -y
apt install -y python-pip unzip
curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip"
unzip awscli-bundle.zip
sudo ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws
mkdir -p /opt/sgt
aws s3 cp s3://${bucket_name}/sgt/sgt /opt/sgt/sgt
aws s3 cp s3://${bucket_name}/sgt/config.json /opt/sgt/config.json
aws s3 cp s3://${bucket_name}/sgt/fullchain.pem /opt/sgt/fullchain.pem
aws s3 cp s3://${bucket_name}/sgt/privkey.pem /opt/sgt/privkey.pem
chmod +x /opt/sgt/sgt
cd /opt/sgt
sleep 30
/opt/sgt/sgt server
