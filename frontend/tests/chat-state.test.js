import test from 'node:test'
import assert from 'node:assert/strict'

import {
  createDraftChatState,
  resolveConversationListState,
  shouldCreateConversationBeforeSending
} from '../src/stores/chat-state.js'

test('createDraftChatState enters a clean unsaved draft conversation', () => {
  const state = createDraftChatState()

  assert.deepEqual(state, {
    activeConversationId: null,
    isDraftConversation: true,
    messagesMap: {}
  })
})

test('resolveConversationListState preserves draft mode while refreshing history list', () => {
  const list = [{ id: 'history-1' }, { id: 'history-2' }]

  const state = resolveConversationListState({
    list,
    activeConversationId: 'history-1',
    messagesMap: { 'history-1': [{ id: 'm1' }] },
    startFresh: true
  })

  assert.deepEqual(state, {
    activeConversationId: null,
    isDraftConversation: true,
    messagesMap: {}
  })
})

test('resolveConversationListState restores first history conversation during normal entry', () => {
  const list = [{ id: 'history-1' }, { id: 'history-2' }]

  const state = resolveConversationListState({
    list,
    activeConversationId: null,
    messagesMap: {},
    startFresh: false
  })

  assert.deepEqual(state, {
    activeConversationId: 'history-1',
    isDraftConversation: false,
    messagesMap: {}
  })
})

test('resolveConversationListState keeps the current active conversation when it still exists', () => {
  const list = [{ id: 'history-1' }, { id: 'history-2' }]

  const state = resolveConversationListState({
    list,
    activeConversationId: 'history-2',
    messagesMap: { 'history-2': [{ id: 'm1' }] },
    startFresh: false
  })

  assert.deepEqual(state, {
    activeConversationId: 'history-2',
    isDraftConversation: false,
    messagesMap: { 'history-2': [{ id: 'm1' }] }
  })
})

test('shouldCreateConversationBeforeSending only returns true for draft or empty active conversation', () => {
  assert.equal(shouldCreateConversationBeforeSending({ activeConversationId: null, isDraftConversation: true }), true)
  assert.equal(shouldCreateConversationBeforeSending({ activeConversationId: null, isDraftConversation: false }), true)
  assert.equal(shouldCreateConversationBeforeSending({ activeConversationId: 'history-1', isDraftConversation: false }), false)
})
