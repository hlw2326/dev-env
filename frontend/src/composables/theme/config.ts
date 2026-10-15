import type { ThemeColor, ThemeState } from './types'

// 色彩模式本地缓存键
export const STORAGE_MODE_KEY = 'devenv_theme_mode'

// 强调色本地缓存键
export const STORAGE_COLOR_KEY = 'devenv_theme_color'

// 默认浅色主题色值
export const DEFAULT_LIGHT_COLOR = '#334155'

// 默认深色主题自适应色值
export const DEFAULT_DARK_COLOR = '#3b82f6'

// 默认主题配置
export const DEFAULT_THEME: ThemeState = {
	mode: 'light',
	color: DEFAULT_LIGHT_COLOR,
}

// 预设强调色板
export const THEME_COLORS: ThemeColor[] = [
	{ name: '灰蓝', value: '#334155' },
	{ name: '靛青', value: '#4f46e5' },
	{ name: '蔚蓝', value: '#2563eb' },
	{ name: '翡翠', value: '#059669' },
	{ name: '天青', value: '#0284c7' },
	{ name: '玫瑰', value: '#e11d48' },
	{ name: '琥珀', value: '#d97706' },
	{ name: '紫罗兰', value: '#7c3aed' },
]
