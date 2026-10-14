import { ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import {
	loadConfig,
	saveConfig,
	readConfig,
	writeConfig,
	listExtensions,
	setExtension,
	loadSoftwareLogs,
	stopSoftware,
	startSoftware,
} from '@/api'
import type { ConfigItem, Extension } from '@/types'

export function useConfig() {
	const toast = useToast()
	const configItems = ref<ConfigItem[]>([])
	const configFile = ref('')
	const rawText = ref('')
	const extensions = ref<Extension[]>([])
	const logLines = ref<string[]>([])
	const saving = ref(false)

	async function fetchConfig(id: number) {
		try {
			const res = await loadConfig(id)
			configItems.value = res.items
			configFile.value = res.file
		} catch (err) {
			toast.add({ severity: 'error', summary: '读取配置失败', detail: String(err), life: 3000 })
		}
	}

	async function submitConfig(id: number, restartNow: boolean, isRunning: boolean) {
		saving.value = true
		try {
			const map: Record<string, string> = {}
			for (const item of configItems.value) {
				map[item.key] = item.value
			}
			const res = await saveConfig({ id, items: map })
			if (!res.ok) {
				toast.add({ severity: 'error', summary: '保存失败', detail: res.message, life: 3500 })
				return
			}
			if (restartNow && isRunning) {
				await stopSoftware(id)
				await startSoftware(id)
				toast.add({ severity: 'success', summary: '已保存并重启', detail: '新配置已生效', life: 2500 })
			} else {
				toast.add({ severity: 'success', summary: '已保存', detail: res.message, life: 2500 })
			}
		} finally {
			saving.value = false
		}
	}

	async function fetchRawFile(id: number) {
		try {
			rawText.value = await readConfig(id)
		} catch (err) {
			toast.add({ severity: 'error', summary: '读取配置文件失败', detail: String(err), life: 3000 })
		}
	}

	async function submitRawFile(id: number) {
		saving.value = true
		try {
			const res = await writeConfig(id, rawText.value)
			if (!res.ok) {
				toast.add({ severity: 'error', summary: '写入失败', detail: res.message, life: 3500 })
				return
			}
			toast.add({ severity: 'success', summary: '已保存', detail: res.message, life: 2000 })
		} finally {
			saving.value = false
		}
	}

	async function fetchExtensions(id: number) {
		try {
			extensions.value = await listExtensions(id)
		} catch (err) {
			toast.add({ severity: 'error', summary: '读取扩展列表失败', detail: String(err), life: 3000 })
		}
	}

	async function toggleExtension(id: number, name: string, on: boolean) {
		const res = await setExtension(id, name, on)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '切换扩展失败', detail: res.message, life: 3000 })
			return
		}
		await fetchExtensions(id)
	}

	async function fetchLogs(id: number) {
		try {
			logLines.value = await loadSoftwareLogs(id, 200)
		} catch (err) {
			toast.add({ severity: 'error', summary: '读取日志失败', detail: String(err), life: 3000 })
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
		fetchRawFile,
		submitRawFile,
		fetchExtensions,
		toggleExtension,
		fetchLogs,
	}
}
