import ToastEventBus from 'primevue/toasteventbus'
import type { ToastMessageOptions } from 'primevue/toast'

// 提示级别类型
export type ToastType = 'info' | 'success' | 'warn' | 'error' | 'secondary' | 'contrast'

// Toast 全局配置
export interface ToastConfig {
	// 持续时长
	duration?: number
	// 分组标识
	group?: string
	// 是否单条覆盖
	single?: boolean
}

// 默认配置对象
let config: ToastConfig = { duration: 2500, single: true }

// 设置 Toast 全局配置
export function setupToast(options: ToastConfig): void {
	config = { ...config, ...options }
}

// 清除所有或指定分组提示
export function clearToast(group?: string): void {
	if (group) {
		ToastEventBus.emit('remove-group', group)
	} else {
		ToastEventBus.emit('remove-all-groups')
	}
}

// 发送一条 Toast 提示
function show(options: ToastMessageOptions): void {
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
		group,
	})
}

// 生成快捷级别函数
function makeLevel(severity: ToastType) {
	return (message: string, detail?: string, life?: number): void => {
		show({ severity, summary: message, detail, life })
	}
}

// 全局统一 Toast 调用函数
export const toast = Object.assign(
	(input: string | ToastMessageOptions, type?: ToastType): void => {
		if (typeof input === 'string') {
			let severity: ToastType = 'info'
			if (type) {
				severity = type
			}
			show({ severity, summary: input })
		} else {
			show(input)
		}
	},
	{
		info: makeLevel('info'),
		success: makeLevel('success'),
		warn: makeLevel('warn'),
		error: makeLevel('error'),
		clear: clearToast,
	},
)

export default toast
