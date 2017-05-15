#!/usr/bin/env bash

# This script is intended to be used to update the API Server running on an
# AKTVE server instance. It is likely called by a cron job, but can also be
# called manually if necessary.

# Stop the server
systemctl stop aktveapisvr
systemctl disable aktveapisvr

# Update the server
su -s /bin/sh gitadmin-ta-aktveapiserver -c 'cd /opt/TA-AKTVEAPIServer && git pull && cd -'
rsync -zvh /opt/TA-AKTVEAPIServer/aktveapisvr.service /etc/systemd/system/aktveapisvr.service
export GOPATH=/opt/go && export GOBIN=$GOPATH/bin && export PATH=$PATH:/usr/local/go/bin:$GOBIN && cd /opt/TA-AKTVEAPIServer && go get && go install && go build -o /opt/TA-AKTVEAPIServer/ta-aktveapiserver && cd -

# Restart the server
systemctl enable aktveapisvr
systemctl start aktveapisvr
