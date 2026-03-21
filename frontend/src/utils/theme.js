export const getSystemTheme = () => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const getCurrentTheme = () => {
  return document.documentElement.dataset.theme || localStorage.getItem('theme') || 'light'
}

export const applyTheme = (theme) => {
  const nextTheme = theme === 'dark' ? 'dark' : 'light'
  document.documentElement.dataset.theme = nextTheme
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
