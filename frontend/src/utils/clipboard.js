// Clipboard helper with a fallback for non-secure contexts: the app is often
// served over plain http://IP:PORT, where navigator.clipboard is unavailable.
export const copyText = async (text) => {
  if (!text) {
    return false
  }
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch (error) {
      // fall through to the legacy path
    }
  }
  return copyViaExecCommand(text)
}

const copyViaExecCommand = (text) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  textarea.setSelectionRange(0, text.length)
  let copied = false
  try {
    copied = document.execCommand('copy')
  } catch (error) {
    copied = false
  }
  document.body.removeChild(textarea)
  return copied
}
