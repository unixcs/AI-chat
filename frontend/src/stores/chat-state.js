export const createDraftChatState = () => ({
  activeConversationId: null,
  isDraftConversation: true,
  messagesMap: {}
})

export const resolveConversationListState = ({ list, activeConversationId, messagesMap, startFresh = false }) => {
  if (startFresh) {
    return createDraftChatState()
  }

  if (!list.length) {
    return {
      activeConversationId: null,
      isDraftConversation: true,
      messagesMap: {}
    }
  }

  const currentStillExists = list.some((item) => item.id === activeConversationId)
  return {
    activeConversationId: currentStillExists ? activeConversationId : list[0].id,
    isDraftConversation: false,
    messagesMap
  }
}

export const shouldCreateConversationBeforeSending = ({ activeConversationId, isDraftConversation }) => {
  return !activeConversationId || isDraftConversation
}
