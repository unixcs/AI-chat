import test from 'node:test'
import assert from 'node:assert/strict'

import { appendAssistantDelta, removeEmptyAssistantPlaceholder } from '../src/stores/chat-streaming.js'

test('appendAssistantDelta writes incoming delta onto the stored assistant message', () => {
  const messages = [
    { id: 'user-1', role: 'user', content: 'hello' },
    { id: 'assistant-1', role: 'assistant', content: '' }
  ]

  appendAssistantDelta(messages, 'assistant-1', '你')
  appendAssistantDelta(messages, 'assistant-1', '好')

  assert.equal(messages[1].content, '你好')
})

test('appendAssistantDelta ignores unknown message ids', () => {
  const messages = [{ id: 'assistant-1', role: 'assistant', content: '' }]

  appendAssistantDelta(messages, 'assistant-404', 'x')

  assert.equal(messages[0].content, '')
})

test('removeEmptyAssistantPlaceholder removes only the targeted empty placeholder', () => {
  const messages = [
    { id: 'user-1', role: 'user', content: 'hello' },
    { id: 'assistant-1', role: 'assistant', content: '' },
    { id: 'assistant-2', role: 'assistant', content: 'world' }
  ]

  const removed = removeEmptyAssistantPlaceholder(messages, 'assistant-1')

  assert.equal(removed, true)
  assert.deepEqual(
    messages.map((message) => message.id),
    ['user-1', 'assistant-2']
  )
})

test('removeEmptyAssistantPlaceholder does not delete the latest real message after array replacement', () => {
  const replacedMessages = [
    { id: 'server-user-1', role: 'user', content: 'hello' },
    { id: 'server-assistant-1', role: 'assistant', content: 'real history' }
  ]

  const removed = removeEmptyAssistantPlaceholder(replacedMessages, 'local-assistant-1')

  assert.equal(removed, false)
  assert.deepEqual(replacedMessages, [
    { id: 'server-user-1', role: 'user', content: 'hello' },
    { id: 'server-assistant-1', role: 'assistant', content: 'real history' }
  ])
})
