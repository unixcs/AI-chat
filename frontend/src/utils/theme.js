// 双机制暗色切换（Plan §3.4 评审 F2 修正）：
// 1) data-theme 属性 —— 本应用历史机制，旧 scoped 样式按此驱动；
// 2) html.dark class —— shadcn/tailwind 生态标准变体，ui/ 组件按此驱动。
// 两者始终同步，verify-dark-mode.py 断言并存。
export const getSystemTheme = () => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const getCurrentTheme = () => {
  return document.documentElement.dataset.theme || localStorage.getItem('theme') || 'light'
}

export const applyTheme = (theme) => {
  const nextTheme = theme === 'dark' ? 'dark' : 'light'
  document.documentElement.dataset.theme = nextTheme
  document.documentElement.classList.toggle('dark', nextTheme === 'dark')
  localStorage.setItem('theme', nextTheme)
  return nextTheme
}

export const initTheme = () => {
  const savedTheme = localStorage.getItem('theme')
  const theme = savedTheme || 'dark'
  return applyTheme(theme)
}

export const toggleTheme = () => {
  const currentTheme = getCurrentTheme()
  return applyTheme(currentTheme === 'dark' ? 'light' : 'dark')
}
