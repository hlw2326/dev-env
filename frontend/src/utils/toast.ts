import ToastEventBus from 'primevue/toasteventbus'
import type { ToastMessageOptions } from 'primevue/toast'

// 提示级别类型
export type ToastType = 'info' | 'success' | 'warn' | 'error' | 'secondary' | 'contrast'

// Toast 全局配置
export interface ToastConfig {
	duration?: number
	group?: string
	single?: boolean
}

// 默认配置对象
let config: ToastConfig = { duration: 2500, single: true }

// 设置 Toast 全局配置
export function setupToast(options: ToastConfig = {}) {
	config = { ...config, ...options }
}

// 清除所有或指定分组提示
export function clearToast(group?: string) {
	if (group) {
		ToastEventBus.emit('remove-group', group)
	} else {
		ToastEventBus.emit('remove-all-groups')
	}
}

// 发送一条 Toast 提示
function show(options: ToastMessageOptions) {
	let group = config.group
	if (options.group) {
		group = options.group
	}
	if (config.single) {
		clearToast(group)
	}
	let severity: ToastMessageOptions['severity'] = 'info'
	if (options.severity) {
		severity = options.severity
	}
	let life = config.duration
	if (options.life) {
		life = options.life
	}
	ToastEventBus.emit('add', {
		closable: false,
		...options,
		severity,
		life,
		group
	})
}

// 生成快捷级别函数
const level = (severity: ToastType) => (message: string, detail?: string, life?: number) => {
	show({ severity, summary: message, detail, life })
}

// 全局统一 Toast 调用函数
export const toast = Object.assign(
	(msgOrOpts: string | ToastMessageOptions, type: ToastType = 'info') => {
		if (typeof msgOrOpts === 'string') {
			show({ severity: type, summary: msgOrOpts })
		} else {
			show(msgOrOpts)
		}
	},
	{
		info: level('info'),
		success: level('success'),
		warn: level('warn'),
		error: level('error'),
		clear: clearToast
	}
)

export default toast
