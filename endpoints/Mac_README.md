### Mac installation and configuration

#### default behavior change

by default, the mac client uses the osquery.flags.default file for its config. NOT osquery.flags



```bash

#!/bin/bash
curl https://pkg.osquery.io/deb/osquery-dbg_2.9.0_1.linux.amd64.debg > osquery.pkg
sudo installer -pkg osquery.pkg -target /
sudo ln -s /usr/local/share/osquery /var/osquery
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
# Create osquery.secret
sudo touch /var/osquery/osquery.secret
# Fill osquery.secret with stuff
sudo echo "demo secret string" > /var/osquery/osquery.secret
# Create osquery.secret osquery.flags.default
sudo touch /var/osquery/osquery.flags.default
# Fill osquery.flags.default with stuff
sudo echo "--config_plugin=tls
--enroll_secret_path=/var/osquery/osquery.secret
--enroll_tls_endpoint=/node/enroll
--config_tls_endpoint=/node/configure
--tls_hostname=<your tls hostname here>
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

```


