const FRESH_CHAT_ON_NEXT_ENTRY_KEY = 'freshChatOnNextEntry'
const DRAFT_CHAT_SESSION_KEY = 'draftChatSessionActive'

export const storeFreshChatFlag = (storage = localStorage) => {
  storage.setItem(FRESH_CHAT_ON_NEXT_ENTRY_KEY, '1')
}

export const consumeFreshChatFlag = (storage = localStorage) => {
  const shouldStartFresh = storage.getItem(FRESH_CHAT_ON_NEXT_ENTRY_KEY) === '1'
  storage.removeItem(FRESH_CHAT_ON_NEXT_ENTRY_KEY)
  return shouldStartFresh
}

export const storeDraftSessionFlag = (storage = sessionStorage) => {
  storage.setItem(DRAFT_CHAT_SESSION_KEY, '1')
}

export const hasDraftSessionFlag = (storage = sessionStorage) => {
  return storage.getItem(DRAFT_CHAT_SESSION_KEY) === '1'
}

export const clearDraftSessionFlag = (storage = sessionStorage) => {
  storage.removeItem(DRAFT_CHAT_SESSION_KEY)
}

export const shouldCreateFreshConversationOnEntry = (hasFreshChatFlag) => {
  return hasFreshChatFlag
}

export const shouldStartFreshOnChatEntry = ({ hasFreshChatFlag, hasDraftSession }) => {
  return Boolean(hasFreshChatFlag || hasDraftSession)
}
