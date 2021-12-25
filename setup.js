const { spawnSync } = require('child_process')

if (process.env.DISABLE_POST_INSTALL === 'yes') {
  process.exit(0)
} else {
  const p = spawnSync('yarn', ['remix', 'setup', 'node'], { stdio: 'inherit' })
  process.exit(p.status)
}
