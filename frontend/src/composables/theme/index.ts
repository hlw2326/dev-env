import { reactive } from 'vue'
import { applyTheme } from './apply'
import {
	STORAGE_MODE_KEY,
	STORAGE_COLOR_KEY,
	DEFAULT_THEME,
	THEME_COLORS,
} from './config'
import type { ThemeMode, ThemeState } from './types'

export * from './types.d'
export * from './config'
export * from './apply'

// 读取本地缓存色彩模式
function loadMode(): ThemeMode {
	const saved = localStorage.getItem(STORAGE_MODE_KEY)
	if (saved === 'light' || saved === 'dark' || saved === 'system') {
		return saved
	}
	return DEFAULT_THEME.mode
}

// 读取本地缓存强调色
function loadColor(): string {
	const saved = localStorage.getItem(STORAGE_COLOR_KEY)
	if (saved) {
		return saved
	}
	return DEFAULT_THEME.color
}

// 全局响应式主题状态
const themeState = reactive<ThemeState>({
	mode: loadMode(),
	color: loadColor(),
})

// 设置主题色彩模式
export function setMode(mode: ThemeMode): void {
	themeState.mode = mode
	localStorage.setItem(STORAGE_MODE_KEY, mode)
	applyTheme(themeState)
}

// 设置主题强调色
export function setColor(color: string): void {
	themeState.color = color
	localStorage.setItem(STORAGE_COLOR_KEY, color)
	applyTheme(themeState)
}

// 初始化全局主题
export function initTheme(): void {
	applyTheme(themeState)
}

// 主题管理组合式函数
export function useTheme() {
	return {
		theme: themeState,
		setMode,
		setColor,
		initTheme,
		THEME_COLORS,
	}
}

export default useTheme
