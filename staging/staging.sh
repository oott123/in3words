#!/bin/bash
yarn build
source ./staging.env
rsync --exclude '.yarn/cache' --exclude '.yarn/unplugged' --exclude '.yarn/install-state.gz' -avP yarn.lock package.json build public .yarnrc.yml .yarn setup.js "$I3W_REMOTE_HOST":/var/www/in3words/
ssh "$I3W_REMOTE_HOST" < ./staging/install.sh
