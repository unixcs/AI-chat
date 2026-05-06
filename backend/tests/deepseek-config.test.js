const test = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')

const backendDir = path.resolve(__dirname, '..')

test('backend env example defaults to DeepSeek v4 flash', () => {
  const envExample = fs.readFileSync(path.join(backendDir, '.env.example'), 'utf8')

  assert.match(envExample, /^DEEPSEEK_MODEL=deepseek-v4-flash$/m)
  assert.match(envExample, /^DEEPSEEK_BASE_URL=https:\/\/api\.deepseek\.com$/m)
})
