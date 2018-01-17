#!/bin/bash

# Create symbolic link between /usr/local/share/osquery and /var/osquery
sudo ln -s /usr/local/share/osquery /var/osquery

# Set path to osquery log file directory
osqueryLogFolder=`/var/log/osquery/`

# Create osquery log file directory
if [ -d "'$osqueryLogFolder" ]; then
	echo "Directory exists..."
else
	sudo mkdir /var/log/osquery
fi

#remove old secret if it exists
if [ -f "/var/osquery/osquery.secret" ]; then
	echo "secret exists..."
    sudo rm /var/osquery/osquery.secret
fi
# Create osquery.secret
sudo touch /var/osquery/osquery.secret

# Fill osquery.secret with stuff
sudo echo "example-secret" > /var/osquery/osquery.secret

#remove old flags if exists
if [ -f "/var/osquery/osquery.secret" ]; then
	echo "flags exists..."
    sudo rm /var/osquery/osquery.flags.default
fi

# Create osquery.secret osquery.flags.default
sudo touch /var/osquery/osquery.flags.default

# Fill osquery.flags.default with stuff
sudo echo "--config_plugin=tls
--enroll_secret_path=/var/osquery/osquery.secret
--enroll_tls_endpoint=/node/enroll
--config_tls_endpoint=/node/configure
--tls_hostname=example.domain.endpoint.com
--config_refresh=300
--config_tls_accelerated_refresh=300
--config_tls_max_attempts=9999"> /var/osquery/osquery.flags.default

# Create symbolic link from /var/osquery/osquery.flags.default to /var/osquery/osquery.flags
sudo ln -s /var/osquery/osquery.flags.default /var/osquery/osquery.flags

# Copy /var/osquery/com.facebook.osqueryd.plist to /Library/LaunchDaemons/
sudo cp /var/osquery/com.facebook.osqueryd.plist /Library/LaunchDaemons/

# Start osquery
sudo launchctl load /Library/LaunchDaemons/com.facebook.osqueryd.plist

exit 0