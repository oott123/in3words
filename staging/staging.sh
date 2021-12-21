#!/bin/bash
yarn build
source ./staging.env
rsync -avP yarn.lock package.json build public "$I3W_REMOTE_HOST":/var/www/in3words/
ssh "$I3W_REMOTE_HOST" < ./staging/install.sh
