import test from 'node:test'
import assert from 'node:assert/strict'

import {
  consumeFreshChatFlag,
  shouldCreateFreshConversationOnEntry,
  storeFreshChatFlag
} from '../src/utils/chat-entry.js'

const createMemoryStorage = () => {
  const data = new Map()

  return {
    getItem(key) {
      return data.has(key) ? data.get(key) : null
    },
    setItem(key, value) {
      data.set(key, String(value))
    },
    removeItem(key) {
      data.delete(key)
    }
  }
}

test('storeFreshChatFlag marks the next chat entry as requiring a clean draft view', () => {
  const storage = createMemoryStorage()

  storeFreshChatFlag(storage)

  assert.equal(storage.getItem('freshChatOnNextEntry'), '1')
})

test('consumeFreshChatFlag returns true once and clears the marker', () => {
  const storage = createMemoryStorage()
  storage.setItem('freshChatOnNextEntry', '1')

  assert.equal(consumeFreshChatFlag(storage), true)
  assert.equal(storage.getItem('freshChatOnNextEntry'), null)
  assert.equal(consumeFreshChatFlag(storage), false)
})

test('shouldCreateFreshConversationOnEntry only triggers when the login marker is present', () => {
  assert.equal(shouldCreateFreshConversationOnEntry(true), true)
  assert.equal(shouldCreateFreshConversationOnEntry(false), false)
})
