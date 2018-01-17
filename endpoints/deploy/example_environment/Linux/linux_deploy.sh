#!/bin/bash

apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 1484120AC4E9F8A1A577AEEE97A80C63C9D8B80B
add-apt-repository "deb [arch=amd64] https://osquery-packages.s3.amazonaws.com/deb deb main"
apt-get update
apt-get install osquery -y


if [ -f "/etc/osquery/osquery.secret" ]; then
	echo "File exists, removing..."
    sudo rm /etc/osquery/osquery.secret
fi
# Create osquery.secret
sudo touch /etc/osquery/osquery.secret

# Fill osquery.secret with stuff
sudo echo "example-secret" > /etc/osquery/osquery.secret

#remove old flags if exists
if [ -f "/etc/osquery/osquery.flags.default" ]; then
	echo "File exists, removing..."
    sudo rm /etc/osquery/osquery.flags.default
fi

# Create osquery.secret osquery.flags.default
sudo touch /etc/osquery/osquery.flags.default

# Fill osquery.flags.default with stuff
sudo echo "--config_plugin=tls
--enroll_secret_path=/etc/osquery/osquery.secret
--enroll_tls_endpoint=/node/enroll
--config_tls_endpoint=/node/configure
--tls_hostname=example.domain.endpoint.com
--config_refresh=300
--config_tls_accelerated_refresh=300
--config_tls_max_attempts=9999"> /etc/osquery/osquery.flags.default

# Create symbolic link from /etc/osquery/osquery.flags.default to /etc/osquery/osquery.flags
if [ -f "/etc/osquery/osquery.flags" ]; then
	echo "File exists, removing..."
    sudo rm /etc/osquery/osquery.flags
fi
sudo ln -s /etc/osquery/osquery.flags.default /etc/osquery/osquery.flags

sudo service osqueryd start