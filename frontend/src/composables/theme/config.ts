import type { ThemeColorItem, ThemeState } from './types'

export const STORAGE_MODE_KEY = 'devenv_theme_mode'
export const STORAGE_COLOR_KEY = 'devenv_theme_color'

export const DEFAULT_THEME: ThemeState = {
  mode: 'light',
  color: '#334155'
}

/**
 * 经典与现代化预设强调色板
 */
export const THEME_COLORS: ThemeColorItem[] = [
  { name: '灰蓝', value: '#334155' },
  { name: '靛青', value: '#4f46e5' },
  { name: '蔚蓝', value: '#2563eb' },
  { name: '翡翠', value: '#059669' },
  { name: '天青', value: '#0284c7' },
  { name: '玫瑰', value: '#e11d48' },
  { name: '琥珀', value: '#d97706' },
  { name: '紫罗兰', value: '#7c3aed' },
]
