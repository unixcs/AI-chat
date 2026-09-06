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
//
// R13 真机补充（用户在 yun1 复测仍失败）：iOS Safari / 微信内置浏览器对
// textarea 的 programmatic selection + execCommand('copy') 仍有独立缺陷
// （readonly/焦点/选区时序），即使按 textarea 生命周期写也可能静默失败。
// 业界标准解法（clipboard-polyfill 同款）：contentEditable <span> +
// Range 选区 —— span 的 textContent 产生真实子节点，selectNodeContents
// 得到非空 range。iOS/iPadOS 直接走 span 路径；其余平台 textarea 失败后
// 降级到 span。

const isAppleMobile = () => {
  if (typeof navigator === 'undefined' || !navigator.userAgent) {
    return false
  }
  const ua = navigator.userAgent
  // iPadOS 13+ 伪装桌面 UA，靠触摸点数识别
  return /iP(hone|ad|od)/.test(ua) || (/Macintosh/.test(ua) && navigator.maxTouchPoints > 1)
}

// 通用路径：textarea（桌面 Chrome/Firefox/Android Chrome 实测可靠）
const copyViaTextarea = (text) => {
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

// iOS/兜底路径：contentEditable span + Range（span 有真实子节点，range 非空）
const copyViaSpan = (text) => {
  const span = document.createElement('span')
  span.textContent = text
  span.contentEditable = 'true'
  span.style.whiteSpace = 'pre'
  span.style.position = 'fixed'
  // 视口内（iOS 只复制视口内选区），横向移出屏幕
  span.style.top = '0'
  span.style.left = '-9999px'
  span.style.opacity = '0'
  document.body.appendChild(span)
  const selection = document.getSelection()
  const prevRange =
    selection &&
    typeof selection.rangeCount === 'number' &&
    selection.rangeCount > 0 &&
    typeof selection.getRangeAt === 'function'
      ? selection.getRangeAt(0)
      : null
  let copied = false
  try {
    if (selection && typeof document.createRange === 'function') {
      const range = document.createRange()
      range.selectNodeContents(span)
      if (typeof selection.removeAllRanges === 'function') {
        selection.removeAllRanges()
      }
      if (typeof selection.addRange === 'function') {
        selection.addRange(range)
      }
    }
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
  document.body.removeChild(span)
  return copied
}

const copyViaExecCommand = (text) => {
  // iOS/iPadOS（含微信内置浏览器的 WKWebView）：textarea 路径有静默失败史，
  // 直接走业界验证过的 span+Range 方案。
  if (isAppleMobile()) {
    return copyViaSpan(text)
  }
  if (copyViaTextarea(text)) {
    return true
  }
  return copyViaSpan(text)
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
