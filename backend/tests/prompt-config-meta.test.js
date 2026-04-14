const test = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')

test('prompt config regression test does not depend on backend .env', () => {
  const testFile = fs.readFileSync(path.join(__dirname, 'prompt-config.test.js'), 'utf8')

  assert.doesNotMatch(testFile, /readText\('\.env'\)/)
})
