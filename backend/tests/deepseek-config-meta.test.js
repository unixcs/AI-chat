const test = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')

const repoRoot = path.resolve(__dirname, '..', '..')

const readRepoFile = (relativePath) => {
  return fs.readFileSync(path.join(repoRoot, relativePath), 'utf8')
}

test('repo docs and backend defaults no longer reference deprecated DeepSeek aliases', () => {
  const backendServer = readRepoFile('backend/server.js')
  const readme = readRepoFile('README.md')
  const quickReference = readRepoFile('PROJECT_QUICK_REFERENCE.md')

  for (const text of [backendServer, readme, quickReference]) {
    assert.doesNotMatch(text, /deepseek-chat/)
    assert.doesNotMatch(text, /https:\/\/api\.deepseek\.com\/v1/)
  }
})
