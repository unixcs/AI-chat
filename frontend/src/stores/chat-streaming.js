export const appendAssistantDelta = (messages, messageId, delta) => {
  if (!Array.isArray(messages) || !messageId || !delta) {
    return false
  }

  const targetMessage = messages.find((message) => message.id === messageId)
  if (!targetMessage) {
    return false
  }

  targetMessage.content += delta
  return true
}

export const removeEmptyAssistantPlaceholder = (messages, messageId) => {
  if (!Array.isArray(messages) || !messageId) {
    return false
  }

  const targetIndex = messages.findIndex((message) => message.id === messageId)
  if (targetIndex < 0) {
    return false
  }

  if (messages[targetIndex].content) {
    return false
  }

  messages.splice(targetIndex, 1)
  return true
}
