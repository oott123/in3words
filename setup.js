const { spawnSync } = require('child_process')

if (process.env.DISABLE_POST_INSTALL === 'yes') {
  process.exit(0)
} else {
  spawnSync('yarn', ['remix', 'setup', 'node'], { stdio: 'inherit' })
}
