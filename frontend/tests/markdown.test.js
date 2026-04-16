import test from 'node:test'
import assert from 'node:assert/strict'
import { JSDOM } from 'jsdom'

test('markdown module can be imported without a browser window', async () => {
  const previousWindow = globalThis.window
  const previousDocument = globalThis.document

  delete globalThis.window
  delete globalThis.document

  try {
    const module = await import(`../src/utils/markdown.js?import-safe=${Date.now()}`)

    assert.equal(typeof module.renderMarkdownToSafeHtml, 'function')
  } finally {
    globalThis.window = previousWindow
    globalThis.document = previousDocument
  }
})

const { window } = new JSDOM('')

globalThis.window = window
globalThis.document = window.document

const { renderMarkdownToSafeHtml } = await import(`../src/utils/markdown.js?browser-env=${Date.now()}`)

test('renderMarkdownToSafeHtml renders emphasis and paragraphs', () => {
  const html = renderMarkdownToSafeHtml('**你好**\n\n第二段')

  assert.match(html, /<strong>你好<\/strong>/)
  assert.match(html, /<p>第二段<\/p>/)
})

test('renderMarkdownToSafeHtml renders markdown lists', () => {
  const html = renderMarkdownToSafeHtml('* 一\n* 二')

  assert.match(html, /<ul>/)
  assert.match(html, /<li>一<\/li>/)
  assert.match(html, /<li>二<\/li>/)
})

test('renderMarkdownToSafeHtml renders fenced code blocks', () => {
  const html = renderMarkdownToSafeHtml('```js\nconst answer = 42\n```')

  assert.match(html, /<pre><code class="language-js">/)
  assert.match(html, /const answer = 42/)
})

test('renderMarkdownToSafeHtml renders markdown tables', () => {
  const html = renderMarkdownToSafeHtml('| 标题 | 说明 |\n| --- | --- |\n| 一行 | 超级超级超级超级超级超级长的内容 |')

  assert.match(html, /<table>/)
  assert.match(html, /<td>超级超级超级超级超级超级长的内容<\/td>/)
})

test('renderMarkdownToSafeHtml strips dangerous html', () => {
  const html = renderMarkdownToSafeHtml('before <script>alert(1)</script> after')

  assert.doesNotMatch(html, /<script/i)
  assert.match(html, /before/)
  assert.match(html, /after/)
})

test('renderMarkdownToSafeHtml hardens links for new tabs', () => {
  const html = renderMarkdownToSafeHtml('[OpenAI](https://openai.com)')

  assert.match(html, /target="_blank"/)
  assert.match(html, /rel="noopener noreferrer"/)
})
