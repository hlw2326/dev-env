import type { ThemeState } from './types'

let mediaQueryListener: ((e: MediaQueryListEvent) => void) | null = null

/**
 * 将主题模式与强调色应用到 DOM 根节点与 CSS 变量
 */
export function applyTheme(state: ThemeState): void {
  const root = document.documentElement
  const isSystemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  const isDark = state.mode === 'dark' || (state.mode === 'system' && isSystemDark)

  root.classList.toggle('dark', isDark)

  const vars = [
    '--p-primary-color',
    '--primary-color',
    '--p-primary-500',
    '--p-primary-hover-color',
    '--p-primary-active-color',
    '--p-button-primary-background',
    '--p-button-primary-border-color',
    '--p-toggleswitch-checked-background',
  ]
  for (const v of vars) {
    root.style.setProperty(v, state.color)
  }

  // 监听操作系统色彩模式变化
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  if (mediaQueryListener) {
    mediaQuery.removeEventListener('change', mediaQueryListener)
    mediaQueryListener = null
  }

  if (state.mode === 'system') {
    mediaQueryListener = (e: MediaQueryListEvent) => {
      root.classList.toggle('dark', e.matches)
    }
    mediaQuery.addEventListener('change', mediaQueryListener)
  }
}
