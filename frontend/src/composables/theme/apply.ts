import { DEFAULT_DARK_COLOR, DEFAULT_LIGHT_COLOR } from './config'
import type { ThemeState } from './types'

// 主题关联的 CSS 变量列表
const PRIMARY_VARS = [
	'--p-primary-color',
	'--primary-color',
	'--p-primary-500',
	'--p-primary-hover-color',
	'--p-primary-active-color',
	'--p-button-primary-background',
	'--p-button-primary-border-color',
	'--p-toggleswitch-checked-background',
]

// 系统色彩模式监听器引用
let mediaHandler: ((event: MediaQueryListEvent) => void) | null = null

// 将主题模式与强调色应用到根节点与 CSS 变量
export function applyTheme(state: ThemeState): void {
	const root = document.documentElement
	const mediaList = window.matchMedia('(prefers-color-scheme: dark)')
	const systemDark = mediaList.matches
	const isDark = state.mode === 'dark' || (state.mode === 'system' && systemDark)

	root.classList.toggle('dark', isDark)

	let activeColor = state.color
	if (isDark && state.color.toLowerCase() === DEFAULT_LIGHT_COLOR) {
		activeColor = DEFAULT_DARK_COLOR
	}

	for (const key of PRIMARY_VARS) {
		root.style.setProperty(key, activeColor)
	}

	if (mediaHandler) {
		mediaList.removeEventListener('change', mediaHandler)
		mediaHandler = null
	}

	if (state.mode === 'system') {
		mediaHandler = (event: MediaQueryListEvent) => {
			root.classList.toggle('dark', event.matches)
		}
		mediaList.addEventListener('change', mediaHandler)
	}
}
