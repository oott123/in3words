#!/bin/bash
set -eo pipefail
cd /var/www/in3words
chown www-data:www-data -R .
sudo -u www-data yarn --production  --ignore-scripts
