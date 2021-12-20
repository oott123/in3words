#!/bin/bash
rsync -avP yarn.lock package.json build public "$I3W_REMOTE_HOST":/var/www/in3words/
ssh "$I3W_REMOTE_HOST" /var/www/in3words/scripts/install.sh
