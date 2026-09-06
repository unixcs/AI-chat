// Clipboard helper with an iOS-safe fallback for non-secure contexts: the app
// is usually served over plain http://IP:PORT, where navigator.clipboard is
// unavailable. Copy must always receive the RAW message text (the Markdown
// source), never rendered HTML.
//
// Ordering constraint (locked by tests): when the Clipboard API is
// unavailable, the execCommand fallback must run BEFORE the first await of
// copyText — user activation dies across microtasks on Safari, so the
// insecure path is kept fully synchronous.

const copyViaExecCommand = (text) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  // NOT readonly: iOS WebKit refuses to select readonly fields. contentEditable
  // keeps iOS from collapsing the selection.
  textarea.contentEditable = 'true'
  textarea.setAttribute('readonly', 'readonly')
  textarea.style.position = 'fixed'
  textarea.style.top = '0'
  textarea.style.left = '-9999px'
  textarea.style.opacity = '0'
  const prevActive = document.activeElement
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()
  textarea.setSelectionRange(0, text.length)
  // Range-API selection on top: the iOS path that actually produces a copyable
  // selection when select()/setSelectionRange silently do nothing.
  const selection = document.getSelection()
  let range = null
  if (selection) {
    selection.removeAllRanges()
    range = document.createRange()
    range.selectNodeContents(textarea)
    selection.addRange(range)
  }
  let copied = false
  try {
    copied = document.execCommand('copy')
  } catch (error) {
    copied = false
  }
  if (selection && range) {
    selection.removeAllRanges()
  }
  document.body.removeChild(textarea)
  if (prevActive && typeof prevActive.focus === 'function') {
    prevActive.focus()
  }
  return copied
}

export const copyText = async (text) => {
  if (typeof text !== 'string' || text === '') {
    return false
  }
  // Clipboard API first — in every context where it is actually exposed.
  if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch (error) {
      // Permission denied / document not focused → legacy path below.
    }
  }
  // Synchronous fallback (see ordering constraint above).
  return copyViaExecCommand(text)
}
