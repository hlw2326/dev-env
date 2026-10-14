import { reactive } from 'vue'
import { applyTheme } from './apply'
import {
  STORAGE_MODE_KEY,
  STORAGE_COLOR_KEY,
  DEFAULT_THEME,
  THEME_COLORS
} from './config'
import type { ThemeMode, ThemeState } from './types'

export * from './types.d'
export * from './config'
export * from './apply'

const initialMode = (localStorage.getItem(STORAGE_MODE_KEY) as ThemeMode) || DEFAULT_THEME.mode
const initialColor = localStorage.getItem(STORAGE_COLOR_KEY) || DEFAULT_THEME.color

// 全局响应式主题状态
const themeState = reactive<ThemeState>({
  mode: initialMode,
  color: initialColor
})

/**
 * 设置色彩模式（浅色/深色/跟随系统）
 */
export function setThemeMode(mode: ThemeMode): void {
  themeState.mode = mode
  localStorage.setItem(STORAGE_MODE_KEY, mode)
  applyTheme(themeState)
}

/**
 * 设置强调色
 */
export function setThemeColor(color: string): void {
  themeState.color = color
  localStorage.setItem(STORAGE_COLOR_KEY, color)
  applyTheme(themeState)
}

/**
 * 全局初始化主题
 */
export function initTheme(): void {
  applyTheme(themeState)
}

/**
 * 主题组合式 Hook
 */
export function useTheme() {
  return {
    theme: themeState,
    setThemeMode,
    setThemeColor,
    initTheme,
    THEME_COLORS
  }
}

export default useTheme
