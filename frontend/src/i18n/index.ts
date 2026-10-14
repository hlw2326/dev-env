import { reactive } from 'vue'
import type { App } from 'vue'
import { enUS } from './locales/en-US'

export { enUS }

// 语言存储键
const STORAGE_KEY = 'devenv_locale'

// 读取本地语言
const stored = localStorage.getItem(STORAGE_KEY)
let initLang = 'zh-CN'
if (stored) {
	initLang = stored
}

// 响应式语言状态
const state = reactive({
	locale: initLang,
})

// 切换语言
export function setLocale(lang: string): void {
	state.locale = lang
	localStorage.setItem(STORAGE_KEY, lang)
}

// 获取当前语言
export function getLocale(): string {
	return state.locale
}

// 翻译文本（中文原文作为键）
export function $t(key: string, params?: Record<string, unknown>): string {
	let text = key
	if (state.locale === 'en-US') {
		const mapped = enUS[key]
		if (mapped) {
			text = mapped
		}
	}
	if (params) {
		for (const name of Object.keys(params)) {
			text = text.replaceAll(`{${name}}`, String(params[name]))
		}
	}
	return text
}

// 语言组合式函数
export function useI18n() {
	return {
		locale: state,
		setLocale,
		getLocale,
		$t,
	}
}

// Vue 插件安装
export const i18n = {
	install(app: App) {
		app.config.globalProperties.$t = $t
	},
}

export default i18n
