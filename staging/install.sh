#!/bin/bash
set -xeo pipefail
cd /var/www/in3words
chown www-data:www-data -R .
sudo -u www-data sh -c 'DISABLE_POST_INSTALL=yes yarn workspaces focus --all --production'
# cp /var/www/in3words/staging/in3words.service /etc/systemd/system/in3words.service
# chown root:root /etc/systemd/system/in3words.service
# chmod 644 /etc/systemd/system/in3words.service
# systemctl daemon-reload
systemctl restart in3words.service
