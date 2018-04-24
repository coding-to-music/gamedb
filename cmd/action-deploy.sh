#!/bin/sh

cd ../

echo "### Pulling"
git fetch origin
git reset --hard origin/master

echo "### Building"
dep ensure
go build

echo "### Copying crontab"
cp ./crontab /etc/cron.d/steamauthority

echo "### Copying varnish VCL"
cp ./varnish.vcl /usr/local/etc/varnish/default.vcl

echo "### Talking to Rollbar"
curl https://api.rollbar.com/api/1/deploy/ \
  -F access_token=${STEAM_ROLLBAR_PRIVATE} \
  -F environment=${ENV} \
  -F revision=$(git log -n 1 --pretty=format:"%H") \
  -F local_username=Jleagle \
  --silent > /dev/null

echo "### Restarting"
/etc/init.d/steam restart
