// Clipboard helper with an iOS-safe fallback for non-secure contexts: the app
// is usually served over plain http://IP:PORT, where navigator.clipboard is
// unavailable. Copy must always receive the RAW message text (the Markdown
// source), never rendered HTML.
//
// Ordering constraint (locked by tests): when the Clipboard API is
// unavailable, the execCommand fallback must run BEFORE the first await of
// copyText — user activation dies across microtasks on Safari, so the
// insecure path is kept fully synchronous.
//
// R12a-P0 regression lock: do NOT touch window.getSelection() between
// select()/setSelectionRange() and execCommand('copy'), and never build a
// Range over the textarea with selectNodeContents() — textareas have no child
// nodes, so that range collapses to empty and removeAllRanges()+addRange()
// destroys the real selection, silently copying nothing (proven in
// Chromium 149 via CDP: execCommand returns true while the OS clipboard keeps
// the stale content).

const copyViaExecCommand = (text) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.contentEditable = 'true'
  // readonly while focusing keeps mobile keyboards from popping up; it is
  // removed before selecting because iOS WebKit refuses to select readonly
  // fields (focus happens while readonly, selection happens after removal).
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'absolute'
  // iOS only copies selections inside the viewport: park at the current
  // scroll position, off-screen horizontally.
  textarea.style.top = `${window.pageYOffset || document.documentElement.scrollTop || 0}px`
  textarea.style.left = '-9999px'
  textarea.style.opacity = '0'
  const prevActive = document.activeElement
  const selection = document.getSelection()
  const prevRange =
    selection &&
    typeof selection.rangeCount === 'number' &&
    selection.rangeCount > 0 &&
    typeof selection.getRangeAt === 'function'
      ? selection.getRangeAt(0)
      : null
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.removeAttribute('readonly')
  textarea.select()
  textarea.setSelectionRange(0, text.length)
  let copied = false
  try {
    copied = document.execCommand('copy')
  } catch (error) {
    copied = false
  }
  if (selection && typeof selection.removeAllRanges === 'function') {
    selection.removeAllRanges()
    if (prevRange && typeof selection.addRange === 'function') {
      selection.addRange(prevRange)
    }
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
