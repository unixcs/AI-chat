import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { JSDOM } from 'jsdom'

const root = resolve(import.meta.dirname, '..')
const source = readFileSync(resolve(root, 'src/utils/clipboard.js'), 'utf8')

// ---------- jsdom 环境 stub ----------

const makeEnv = ({ hasClipboardAPI = false, apiRejects = false, execResult = true } = {}) => {
  const calls = { writeText: [], exec: [], ranges: [] }
  const dom = new JSDOM('<!doctype html><html><body></body></html>', { url: 'http://localhost/' })
  global.window = dom.window
  global.document = dom.window.document
  Object.defineProperty(globalThis, 'navigator', { value: dom.window.navigator, configurable: true })
  if (hasClipboardAPI) {
    Object.defineProperty(global.navigator, 'clipboard', {
      value: {
        writeText: async (t) => {
          if (apiRejects) throw new Error('NotAllowedError')
          calls.writeText.push(t)
        },
      },
      configurable: true,
    })
  }
  const fakeSelection = {
    removeAllRanges: () => calls.ranges.push('removeAllRanges'),
    addRange: (r) => calls.ranges.push(['addRange', typeof r]),
  }
  global.document.getSelection = () => fakeSelection
  global.document.createRange = () => ({ selectNodeContents: () => {} })
  global.document.execCommand = (cmd) => {
    calls.exec.push(cmd)
    return execResult
  }
  Object.defineProperty(global.document, 'activeElement', { value: global.document.body, configurable: true })
  return calls
}

const resetEnv = () => {
  delete global.window
  delete global.document
}

// copyText 读取顶层 navigator/document，须在 stub 后再动态 import
const loadCopyText = async () => (await import(`../src/utils/clipboard.js?t=${Date.now()}-${Math.random()}`)).copyText

test('scenario matrix: copies raw text for user msgs, AI replies, long text, markdown, emoji, multiline, numeric-string', async (t) => {
  for (const [name, text] of Object.entries({
    用户消息: '你好，塔罗解读一下今天运势',
    AI回复: '**今天**运势不错！\n\n- 事业：上升\n- 感情：平稳',
    长文本: '长'.repeat(20000),
    markdown源文: '# 标题\n```js\nconst a = 1\n```\n| a | b |\n|---|---|\n| 1 | 2 |',
    emoji: '🔮✨🌟 牌面开启',
    多行: '第一行\n第二行\n第三行',
    数字字符串: '123456',
  })) {
    await t.test(name, async () => {
      const calls = makeEnv({ hasClipboardAPI: false })
      const copyText = await loadCopyText()
      assert.equal(await copyText(text), true)
      assert.equal(calls.exec.filter((c) => c === 'copy').length, 1)
      resetEnv()
    })
  }
})

test('empty and non-string input is rejected', async () => {
  const calls = makeEnv({ hasClipboardAPI: false })
  const copyText = await loadCopyText()
  assert.equal(await copyText(''), false)
  assert.equal(await copyText(undefined), false)
  assert.equal(await copyText(12345), false)
  assert.equal(calls.exec.length, 0)
  resetEnv()
})

test('Clipboard API available → preferred, called exactly once', async () => {
  const calls = makeEnv({ hasClipboardAPI: true, apiRejects: false })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  assert.deepEqual(calls.writeText, ['内容'])
  assert.equal(calls.exec.length, 0)
  resetEnv()
})

test('Clipboard API rejection → falls back to execCommand path', async () => {
  const calls = makeEnv({ hasClipboardAPI: true, apiRejects: true })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  assert.equal(calls.writeText.length, 0)
  assert.deepEqual(calls.exec, ['copy'])
  resetEnv()
})

test('execCommand returning false → copyText false (failure must surface)', async () => {
  const calls = makeEnv({ hasClipboardAPI: false, execResult: false })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), false)
  assert.deepEqual(calls.exec, ['copy'])
  resetEnv()
})

test('iOS path: Range selection established, textarea removed after copy', async () => {
  const calls = makeEnv({ hasClipboardAPI: false })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  assert.ok(calls.ranges.some((r) => r === 'removeAllRanges'))
  assert.ok(calls.ranges.some((r) => Array.isArray(r) && r[0] === 'addRange'))
  assert.equal(global.document.querySelectorAll('textarea').length, 0)
  resetEnv()
})

test('fallback stays reachable before the first await (user-activation constraint)', () => {
  const body = source.slice(source.indexOf('export const copyText'))
  const apiBranchEnd = body.indexOf('  // Synchronous fallback')
  const beforeFallback = body.slice(0, apiBranchEnd)
  const awaits = [...beforeFallback.matchAll(/await /g)].length
  assert.equal(awaits, 1, 'exactly the clipboard writeText await before fallback')
  assert.match(source, /contentEditable/)
  assert.match(source, /setSelectionRange/)
  assert.match(source, /addRange/)
})
