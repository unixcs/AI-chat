import { defineStore } from 'pinia'
import { createConversation, getConversations, getMessages, sendMessage, streamMessage } from '../api/chat'
import router from '../router'

export const useChatStore = defineStore('chat', {
  state: () => ({
    list: [],
    activeConversationId: null,
    messagesMap: {},
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
      this.loading = false
      this.streaming = false
      this.streamSource = null
      this.uiResetKey += 1
    },
    async fetchConversations() {
      this.loading = true
      try {
        const { data } = await getConversations()
        this.list = data.data
        if (this.list.length === 0) {
          this.activeConversationId = null
          this.messagesMap = {}
          return
        }

        const currentStillExists = this.list.some((item) => item.id === this.activeConversationId)
        if (!currentStillExists) {
          this.activeConversationId = this.list[0].id
        }

        await this.fetchMessages(this.activeConversationId)
      } finally {
        this.loading = false
      }
    },
    async addConversation() {
      const { data } = await createConversation()
      this.list.unshift(data.data)
      this.activeConversationId = data.data.id
      this.messagesMap[data.data.id] = []
    },
    async fetchMessages(conversationId) {
      if (!conversationId) {
        return
      }
      const { data } = await getMessages(conversationId)
      this.messagesMap[conversationId] = data.data
      this.activeConversationId = conversationId
    },
    async postMessage(content) {
      if (!this.activeConversationId) {
        await this.addConversation()
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
      if (!this.activeConversationId) {
        await this.addConversation()
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
      this.streaming = true

      return new Promise((resolve, reject) => {
        this.streamSource = streamMessage(conversationId, content, {
          onDelta: (delta) => {
            assistantMessage.content += delta
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
            if (!assistantMessage.content) {
              this.messagesMap[conversationId].pop()
            }
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
