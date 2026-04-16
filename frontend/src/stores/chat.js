import { defineStore } from 'pinia'
import { createConversation, getConversations, getMessages, sendMessage, streamMessage } from '../api/chat'
import router from '../router'
import { appendAssistantDelta, removeEmptyAssistantPlaceholder } from './chat-streaming'
import { createDraftChatState, resolveConversationListState, shouldCreateConversationBeforeSending } from './chat-state'

export const useChatStore = defineStore('chat', {
  state: () => ({
    list: [],
    activeConversationId: null,
    messagesMap: {},
    isDraftConversation: true,
    loading: false,
    streaming: false,
    streamSource: null,
    uiResetKey: 0
  }),
  actions: {
    resetChatState() {
      if (this.streamSource) {
        this.streamSource.close()
      }
      this.list = []
      this.activeConversationId = null
      this.messagesMap = {}
      this.isDraftConversation = true
      this.loading = false
      this.streaming = false
      this.streamSource = null
      this.uiResetKey += 1
    },
    startDraftConversation() {
      const draftState = createDraftChatState()
      this.activeConversationId = draftState.activeConversationId
      this.messagesMap = draftState.messagesMap
      this.isDraftConversation = draftState.isDraftConversation
      this.uiResetKey += 1
    },
    async fetchConversations(options = {}) {
      this.loading = true
      try {
        const { data } = await getConversations()
        this.list = data.data
        const nextState = resolveConversationListState({
          list: this.list,
          activeConversationId: this.activeConversationId,
          messagesMap: this.messagesMap,
          startFresh: Boolean(options.startFresh)
        })
        this.activeConversationId = nextState.activeConversationId
        this.messagesMap = nextState.messagesMap
        this.isDraftConversation = nextState.isDraftConversation
        if (!this.activeConversationId) {
          return
        }
        await this.fetchMessages(this.activeConversationId)
      } finally {
        this.loading = false
      }
    },
    async ensureActiveConversation() {
      if (!shouldCreateConversationBeforeSending({
        activeConversationId: this.activeConversationId,
        isDraftConversation: this.isDraftConversation
      })) {
        return this.activeConversationId
      }

      const { data } = await createConversation()
      this.list.unshift(data.data)
      this.activeConversationId = data.data.id
      this.messagesMap[data.data.id] = []
      this.isDraftConversation = false
      return data.data.id
    },
    async addConversation() {
      this.startDraftConversation()
    },
    async fetchMessages(conversationId) {
      if (!conversationId) {
        return
      }
      const { data } = await getMessages(conversationId)
      this.messagesMap[conversationId] = data.data
      this.activeConversationId = conversationId
      this.isDraftConversation = false
    },
    async postMessage(content) {
      if (shouldCreateConversationBeforeSending({
        activeConversationId: this.activeConversationId,
        isDraftConversation: this.isDraftConversation
      })) {
        await this.ensureActiveConversation()
      }

      const conversationId = this.activeConversationId
      if (!this.messagesMap[conversationId]) {
        this.messagesMap[conversationId] = []
      }

      this.messagesMap[conversationId].push({
        id: `local-user-${Date.now()}`,
        role: 'user',
        content,
        createdAt: new Date().toISOString()
      })

      const { data } = await sendMessage(conversationId, content)
      this.messagesMap[conversationId].push(data.data)
    },
    async postStreamMessage(content) {
      if (shouldCreateConversationBeforeSending({
        activeConversationId: this.activeConversationId,
        isDraftConversation: this.isDraftConversation
      })) {
        await this.ensureActiveConversation()
      }

      const conversationId = this.activeConversationId
      if (!this.messagesMap[conversationId]) {
        this.messagesMap[conversationId] = []
      }

      const now = new Date().toISOString()
      this.messagesMap[conversationId].push({
        id: `local-user-${Date.now()}`,
        role: 'user',
        content,
        createdAt: now
      })

      const assistantMessage = {
        id: `local-assistant-${Date.now()}`,
        role: 'assistant',
        content: '',
        createdAt: now
      }
      this.messagesMap[conversationId].push(assistantMessage)
      const assistantMessageId = assistantMessage.id
      this.streaming = true

      return new Promise((resolve, reject) => {
        this.streamSource = streamMessage(conversationId, content, {
          onDelta: (delta) => {
            appendAssistantDelta(this.messagesMap[conversationId], assistantMessageId, delta)
          },
          onDone: () => {
            this.streaming = false
            this.streamSource = null
            this.fetchConversations()
            resolve(true)
          },
          onError: (error) => {
            this.streaming = false
            this.streamSource = null
            const currentMessages = this.messagesMap[conversationId] || []
            removeEmptyAssistantPlaceholder(currentMessages, assistantMessageId)
            if (error.message.includes('账号已在其他设备登录')) {
              localStorage.removeItem('userToken')
              localStorage.removeItem('adminToken')
              this.resetChatState()
              router.push('/login')
            }
            reject(error)
          }
        })
      })
    },
    stopStreaming() {
      if (this.streamSource) {
        this.streamSource.close()
        this.streamSource = null
      }
      this.streaming = false
    }
  }
})
