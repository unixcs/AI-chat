const FRESH_CHAT_ON_NEXT_ENTRY_KEY = 'freshChatOnNextEntry'

export const storeFreshChatFlag = (storage = localStorage) => {
  storage.setItem(FRESH_CHAT_ON_NEXT_ENTRY_KEY, '1')
}

export const consumeFreshChatFlag = (storage = localStorage) => {
  const shouldStartFresh = storage.getItem(FRESH_CHAT_ON_NEXT_ENTRY_KEY) === '1'
  storage.removeItem(FRESH_CHAT_ON_NEXT_ENTRY_KEY)
  return shouldStartFresh
}

export const shouldCreateFreshConversationOnEntry = (hasFreshChatFlag) => {
  return hasFreshChatFlag
}
