import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { JSDOM } from 'jsdom'

const root = resolve(import.meta.dirname, '..')
const source = readFileSync(resolve(root, 'src/utils/clipboard.js'), 'utf8')

// ---------- jsdom 环境 stub ----------
// jsdom 无法证明 OS 剪贴板端到端（R12a 的教训：假绿灯）。这里锁定的是
// 结构不变式（调用顺序/路径分流/Range 覆盖非空内容/失败透出）；真实浏览器
// 的端到端证据由 CDP 工装提供（R12a 12/12 + R13 线上验收）。

const makeEnv = ({
  hasClipboardAPI = false,
  apiRejects = false,
  execResult = true,
  userSelection = null,
  userAgent = 'Mozilla/5.0 (X11; Linux x86_64) Chrome/149.0',
} = {}) => {
  const calls = {
    writeText: [],
    exec: [],
    ranges: [],
    order: [],
    setSelRange: [],
    readonlyRemoved: 0,
    created: [],
    rangeTargets: [],
  }
  const dom = new JSDOM('<!doctype html><html><body></body></html>', { url: 'http://localhost/' })
  global.window = dom.window
  global.document = dom.window.document
  Object.defineProperty(globalThis, 'navigator', { value: dom.window.navigator, configurable: true })
  Object.defineProperty(global.navigator, 'userAgent', { value: userAgent, configurable: true })
  Object.defineProperty(global.navigator, 'maxTouchPoints', { value: userAgent.includes('Macintosh') ? 5 : 0, configurable: true })
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
    rangeCount: userSelection ? 1 : 0,
    getRangeAt: () => userSelection,
    removeAllRanges: () => {
      calls.ranges.push('removeAllRanges')
      calls.order.push('removeAllRanges')
    },
    addRange: (r) => {
      calls.ranges.push(['addRange', r])
      calls.order.push('addRange')
    },
  }
  global.document.getSelection = () => fakeSelection
  // Range stub 记录 selectNodeContents 的目标元素（R12a 教训：必须断言
  // range 覆盖的是非空内容，而不是只断言"调用过"）
  global.document.createRange = () => ({
    selectNodeContents: (el) => calls.rangeTargets.push(el),
  })
  global.document.execCommand = (cmd) => {
    calls.exec.push(cmd)
    calls.order.push(`execCommand:${cmd}`)
    return execResult
  }
  Object.defineProperty(global.document, 'activeElement', { value: global.document.body, configurable: true })

  // 记录 createElement 的元素类型（区分 textarea / span 路径）
  const origCreate = dom.window.document.createElement.bind(dom.window.document)
  dom.window.document.createElement = (tag, ...rest) => {
    const el = origCreate(tag, ...rest)
    calls.created.push(tag.toUpperCase())
    return el
  }

  // 记录 textarea 选区生命周期（jsdom 原生方法静默，须包裹观测）
  const proto = dom.window.HTMLTextAreaElement.prototype
  const wrap = (name) => {
    const orig = proto[name]
    proto[name] = function (...args) {
      calls.order.push(name)
      if (name === 'setSelectionRange') calls.setSelRange.push(args)
      return orig ? orig.apply(this, args) : undefined
    }
  }
  wrap('focus')
  wrap('select')
  wrap('setSelectionRange')
  const origRemoveAttribute = dom.window.Element.prototype.removeAttribute
  dom.window.Element.prototype.removeAttribute = function (name) {
    if (this.tagName === 'TEXTAREA' && name === 'readonly') {
      calls.readonlyRemoved += 1
      calls.order.push('removeAttribute:readonly')
    }
    return origRemoveAttribute.call(this, name)
  }
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

test('execCommand returning false on both paths → copyText false (failure must surface)', async () => {
  const calls = makeEnv({ hasClipboardAPI: false, execResult: false })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), false)
  // textarea 失败后降级到 span 再试一次，全失败才透出 false
  assert.deepEqual(calls.exec, ['copy', 'copy'])
  resetEnv()
})

// R12a-P0 回归锁：选区必须在 select()/setSelectionRange() 建立后、
// execCommand('copy') 前不被任何窗口级选区操作销毁；readonly 在选区前移除
//（iOS readonly 选区缺陷），focus 在选区之前。
test('textarea path lifecycle: focus → readonly removed → select → setSelectionRange → execCommand, window selection untouched until after copy', async () => {
  const calls = makeEnv({ hasClipboardAPI: false })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  const copyIdx = calls.order.indexOf('execCommand:copy')
  assert.ok(copyIdx > -1, 'execCommand copy was issued')
  const lifecycle = calls.order.slice(0, copyIdx)
  assert.deepEqual(lifecycle, ['focus', 'removeAttribute:readonly', 'select', 'setSelectionRange'])
  assert.equal(calls.setSelRange[0]?.[0], 0)
  assert.equal(calls.setSelRange[0]?.[1], '内容'.length)
  assert.equal(calls.readonlyRemoved, 1)
  // 窗口选区只能在 copy 之后被动清理；textarea 路径绝不允许 addRange
  //（textarea 无子节点，selectNodeContents 造出的 range 恒为空——R12a-P0）
  const firstClear = calls.order.indexOf('removeAllRanges')
  assert.ok(firstClear === -1 || firstClear > copyIdx, 'removeAllRanges must not run before execCommand copy')
  assert.equal(calls.ranges.some((r) => Array.isArray(r)), false, 'addRange must never be used on the textarea path')
  // textarea 移除、焦点归还
  assert.equal(global.document.querySelectorAll('textarea').length, 0)
  resetEnv()
})

// R13：iOS/iPadOS（含微信 WKWebView）→ 直接走 contentEditable span + Range
// 路径（span 的 textContent 是真实子节点，selectNodeContents 得到非空 range）
test('Apple mobile UA → span path only: Range must cover the non-empty span content', async () => {
  const calls = makeEnv({
    hasClipboardAPI: false,
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Safari/604.1',
  })
  const copyText = await loadCopyText()
  assert.equal(await copyText('手机复制内容'), true)
  assert.equal(calls.created.includes('TEXTAREA'), false, 'iOS path must not use a textarea')
  assert.equal(calls.created.includes('SPAN'), true)
  assert.equal(calls.rangeTargets.length, 1, 'exactly one Range selection')
  assert.equal(calls.rangeTargets[0].tagName, 'SPAN')
  assert.equal(calls.rangeTargets[0].textContent, '手机复制内容', 'Range must cover the actual text (R12a fake-green guard)')
  assert.equal(calls.setSelRange.length, 0, 'no textarea selection on iOS path')
  assert.equal(global.document.querySelectorAll('span').length, 0)
  resetEnv()
})

test('iPadOS desktop-UA (Macintosh + multi-touch) also routes to span path', async () => {
  const calls = makeEnv({
    hasClipboardAPI: false,
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1.15',
  })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  assert.equal(calls.created.includes('TEXTAREA'), false)
  assert.equal(calls.created.includes('SPAN'), true)
  resetEnv()
})

test('non-Apple: textarea success → span fallback NOT used', async () => {
  const calls = makeEnv({ hasClipboardAPI: false, execResult: true })
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  assert.equal(calls.created.filter((t) => t === 'TEXTAREA').length, 1)
  assert.equal(calls.created.includes('SPAN'), false, 'no span when textarea path succeeds')
  resetEnv()
})

test('non-Apple: textarea execCommand failure → span fallback runs once', async () => {
  let execCount = 0
  const calls = makeEnv({ hasClipboardAPI: false })
  // 第一次 execCommand（textarea）失败，第二次（span）成功
  global.document.execCommand = (cmd) => {
    calls.exec.push(cmd)
    calls.order.push(`execCommand:${cmd}`)
    execCount += 1
    return execCount > 1
  }
  const copyText = await loadCopyText()
  assert.equal(await copyText('内容'), true)
  assert.equal(calls.exec.filter((c) => c === 'copy').length, 2, 'textarea then span')
  assert.equal(calls.created.includes('SPAN'), true)
  resetEnv()
})

test('user selection is preserved and restored after the copy (both paths)', async () => {
  const sentinel = { collapsed: false }
  for (const ua of [
    'Mozilla/5.0 (X11; Linux x86_64) Chrome/149.0',
    'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Safari/604.1',
  ]) {
    const calls = makeEnv({ hasClipboardAPI: false, userSelection: sentinel, userAgent: ua })
    const copyText = await loadCopyText()
    assert.equal(await copyText('内容'), true)
    const restore = calls.ranges.filter((r) => Array.isArray(r) && r[1] === sentinel)
    assert.equal(restore.length, 1, `original range restored exactly once (${ua})`)
    // 恢复动作必须在 copy 之后（span 路径选区建立的那次 addRange 在 copy 前，不算）
    const restoreOrder = calls.order.lastIndexOf('addRange')
    assert.ok(restoreOrder > calls.order.indexOf('execCommand:copy'), ua)
    resetEnv()
  }
})

test('fallback stays reachable before the first await (user-activation constraint)', () => {
  const body = source.slice(source.indexOf('export const copyText'))
  const apiBranchEnd = body.indexOf('  // Synchronous fallback')
  const beforeFallback = body.slice(0, apiBranchEnd)
  const awaits = [...beforeFallback.matchAll(/await /g)].length
  assert.equal(awaits, 1, 'exactly the clipboard writeText await before fallback')
  assert.match(source, /contentEditable/)
  assert.match(source, /setSelectionRange/)
  assert.match(source, /removeAttribute\('readonly'\)/)
  // R12a-P0：破坏性模式被明令禁止——textarea 函数体内不得出现 Range 选区
  //（匹配前剥离注释，锁定的是代码而非说明文字；span 路径的 selectNodeContents
  // 是合法的——span 有真实子节点）
  const taStart = source.indexOf('const copyViaTextarea')
  const taEnd = source.indexOf('const copyViaSpan')
  const taCode = source.slice(taStart, taEnd).replace(/\/\*[\s\S]*?\*\//g, '').replace(/\/\/[^\n]*/g, '')
  assert.doesNotMatch(taCode, /selectNodeContents/, 'selectNodeContents on a textarea builds an empty range (R12a-P0)')
  const spanCode = source.slice(taEnd).replace(/\/\*[\s\S]*?\*\//g, '').replace(/\/\/[^\n]*/g, '')
  assert.match(spanCode, /selectNodeContents\(span\)/, 'span path must Range-select the span contents')
})
