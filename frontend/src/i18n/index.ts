import { reactive } from 'vue'
import type { App } from 'vue'
import { zhCN } from './locales/zh-CN'
import { enUS } from './locales/en-US'

export { zhCN, enUS }

const STORAGE_KEY = 'devenv_locale'
const initialLocale = localStorage.getItem(STORAGE_KEY) || 'zh-CN'

// 响应式语言状态树
const state = reactive({
  locale: initialLocale
})

/**
 * 切换并持久化界面语言
 */
export function setLocale(lang: string): void {
  state.locale = lang
  localStorage.setItem(STORAGE_KEY, lang)
}

/**
 * 获取当前界面语言
 */
export function getLocale(): string {
  return state.locale
}

/**
 * 响应式国际化文本翻译
 */
export function $t(key: string, params?: Record<string, unknown>): string {
  const current = state.locale === 'en-US' ? enUS : zhCN
  let text = current[key] || key
  if (params) {
    for (const name of Object.keys(params)) {
      text = text.replaceAll(`{${name}}`, String(params[name]))
    }
  }
  return text
}

export function useI18n() {
  return {
    locale: state,
    setLocale,
    getLocale,
    $t
  }
}

// Vue 插件
export const i18n = {
  install(app: App) {
    app.config.globalProperties.$t = $t
  }
}

export default i18n
