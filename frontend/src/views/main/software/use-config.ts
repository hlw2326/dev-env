import { ref } from 'vue'
import {
	loadConfig,
	saveConfig,
	readConfig,
	writeConfig,
	listExtensions,
	setExtension,
	listLogs,
	stopSoftware,
	startSoftware,
} from '@/api'
import { toast, formatError } from '@/utils'
import type { ConfigItem, Extension } from '@/types'

// 单次读取最大日志行数
const LOG_LIMIT = 200

// 常用配置项列表
const configItems = ref<ConfigItem[]>([])

// 配置文件路径
const configFile = ref('')

// 配置文件全文内容
const rawText = ref('')

// PHP 扩展列表
const extensions = ref<Extension[]>([])

// 软件运行日志行列表
const logLines = ref<string[]>([])

// 保存中状态
const saving = ref(false)

// 软件配置与日志管理组合式函数
export function useConfig() {
	// 读取软件常用配置项
	async function fetchConfig(id: number): Promise<void> {
		try {
			const result = await loadConfig(id)
			configItems.value = result.items
			configFile.value = result.file
		} catch (err: unknown) {
			toast.error('读取配置失败', formatError(err))
		}
	}

	// 提交保存常用配置项并按需重启服务
	async function submitConfig(id: number, restartNow: boolean, isRunning: boolean): Promise<void> {
		saving.value = true
		try {
			const map: Record<string, string> = {}
			for (const item of configItems.value) {
				map[item.key] = item.value
			}
			const result = await saveConfig({ id, items: map })
			if (!result.ok) {
				toast.error('保存失败', result.message, 3500)
				return
			}
			if (restartNow && result.restart && isRunning) {
				await stopSoftware(id)
				await startSoftware(id)
				toast.success('已保存并重启', '新配置已生效')
			} else {
				toast.success('已保存', result.message)
			}
		} catch (err: unknown) {
			toast.error('保存配置失败', formatError(err))
		} finally {
			saving.value = false
		}
	}

	// 读取配置文件全文
	async function fetchRaw(id: number): Promise<void> {
		try {
			rawText.value = await readConfig(id)
		} catch (err: unknown) {
			toast.error('读取配置文件失败', formatError(err))
		}
	}

	// 保存配置文件全文
	async function submitRaw(id: number): Promise<void> {
		saving.value = true
		try {
			const result = await writeConfig(id, rawText.value)
			if (!result.ok) {
				toast.error('写入失败', result.message, 3500)
				return
			}
			toast.success('已保存', result.message)
		} catch (err: unknown) {
			toast.error('写入配置文件失败', formatError(err))
		} finally {
			saving.value = false
		}
	}

	// 获取 PHP 扩展开关列表
	async function fetchExtensions(id: number): Promise<void> {
		try {
			extensions.value = await listExtensions(id)
		} catch (err: unknown) {
			toast.error('读取扩展列表失败', formatError(err))
		}
	}

	// 切换 PHP 扩展启用状态
	async function toggleExtension(id: number, name: string, on: boolean): Promise<void> {
		try {
			const result = await setExtension(id, name, on)
			if (!result.ok) {
				toast.error('切换扩展失败', result.message)
				return
			}
			await fetchExtensions(id)
		} catch (err: unknown) {
			toast.error('切换扩展失败', formatError(err))
		}
	}

	// 读取软件运行日志末尾行
	async function fetchLogs(id: number): Promise<void> {
		try {
			logLines.value = await listLogs(id, LOG_LIMIT)
		} catch (err: unknown) {
			toast.error('读取日志失败', formatError(err))
		}
	}

	return {
		configItems,
		configFile,
		rawText,
		extensions,
		logLines,
		saving,
		fetchConfig,
		submitConfig,
		fetchRaw,
		submitRaw,
		fetchExtensions,
		toggleExtension,
		fetchLogs,
	}
}
