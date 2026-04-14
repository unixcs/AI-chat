const test = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')

const backendDir = path.resolve(__dirname, '..')

const readText = (fileName) => {
  return fs.readFileSync(path.join(backendDir, fileName), 'utf8')
}

test('backend env example declares the system prompt file setting', () => {
  const envExample = readText('.env.example')

  assert.match(envExample, /^DEEPSEEK_SYSTEM_PROMPT_FILE=prompts\/Prompt\.md$/m)
})
