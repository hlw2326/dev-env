import { computed, ref } from 'vue'
import { listSoftware, listenEvent, restartSoftware, startSoftware, stopSoftware, setMain, setAutostart, openSoftware } from '@/api'
import { formatError, toast } from '@/utils'
import type { ServiceEvent, Software } from '@/types'

// 已安装软件服务列表
const items = ref<Software[]>([])

// 页面加载状态
const loading = ref(false)

// 正在操作的服务标识
const busy = ref(0)

// 当前服务操作方向
const operation = ref('')

// 运行中服务页面组合式函数
export function useServices() {
	// 当前已安装且可管理的服务
	const services = computed(() => {
		return items.value.filter((item) => item.service)
	})

	// 加载服务运行状态
	async function loadData(silent = false): Promise<void> {
		if (!silent) {
			loading.value = true
		}
		try {
			items.value = await listSoftware()
			const current = items.value.find((item) => item.id === busy.value)
			if (current && operation.value === 'start' && current.running) {
				clearBusy()
			}
			if (current && operation.value === 'stop' && !current.running) {
				clearBusy()
			}
			if (current && operation.value === 'restart' && current.running) {
				clearBusy()
			}
		} catch (err: unknown) {
			toast.error('加载服务失败', formatError(err))
		} finally {
			if (!silent) {
				loading.value = false
			}
		}
	}

	// 启动指定服务
	async function startService(id: number): Promise<void> {
		setBusy(id, 'start')
		try {
			const result = await startSoftware(id)
			if (!result.ok) {
				clearBusy()
				toast.error('启动失败', result.message)
				return
			}
			toast.info('服务启动中')
			await loadData()
		} catch (err: unknown) {
			clearBusy()
			toast.error('启动失败', formatError(err))
		}
	}

	// 停止指定服务
	async function stopService(id: number): Promise<void> {
		setBusy(id, 'stop')
		try {
			const result = await stopSoftware(id)
			if (!result.ok) {
				clearBusy()
				toast.error('停止失败', result.message)
				return
			}
			toast.info('服务停止中')
			await loadData()
		} catch (err: unknown) {
			clearBusy()
			toast.error('停止失败', formatError(err))
		}
	}

	// 异步重启指定服务
	async function restartService(id: number): Promise<void> {
		setBusy(id, 'restart')
		try {
			const result = await restartSoftware(id)
			if (!result.ok) {
				clearBusy()
				toast.error('重启失败', result.message)
				return
			}
			toast.info('服务重启中')
			await loadData()
		} catch (err: unknown) {
			clearBusy()
			toast.error('重启失败', formatError(err))
		}
	}

	// 在设置抽屉中启动指定服务
	async function startConfiguredService(item: Software): Promise<void> {
		await startService(item.id)
	}

	// 在设置抽屉中停止指定服务
	async function stopConfiguredService(item: Software): Promise<void> {
		await stopService(item.id)
	}

	// 设置软件默认版本
	async function markMain(item: Software): Promise<void> {
		try {
			const result = await setMain(item.id)
			if (!result.ok) {
				toast.error('设置默认失败', result.message)
				return
			}
			toast.success('已设为默认版本', `${item.kind} ${item.version}`)
			await loadData()
		} catch (err: unknown) {
			toast.error('设置默认失败', formatError(err))
		}
	}

	// 切换服务随应用启动
	async function toggleAutostart(item: Software, on: boolean): Promise<void> {
		try {
			const result = await setAutostart(item.id, on)
			if (!result.ok) {
				toast.error('设置失败', result.message)
				return
			}
			await loadData()
		} catch (err: unknown) {
			toast.error('设置失败', formatError(err))
		}
	}

	// 打开软件安装或数据目录
	async function browseFolder(id: number): Promise<void> {
		try {
			const result = await openSoftware(id)
			if (!result.ok) {
				toast.error('打开目录失败', result.message)
			}
		} catch (err: unknown) {
			toast.error('打开目录失败', formatError(err))
		}
	}

	// 设置服务操作状态
	function setBusy(id: number, target: string): void {
		busy.value = id
		operation.value = target
	}

	// 清理服务操作状态
	function clearBusy(): void {
		busy.value = 0
		operation.value = ''
	}

	// 监听服务状态变化并刷新列表
	function bindEvents(): () => void {
		return listenEvent('service:status-changed', (payload: ServiceEvent) => {
			if (payload.id === busy.value && (operation.value !== 'restart' || payload.running || payload.message !== '')) {
				clearBusy()
			}
			if (payload.message !== '') {
				toast.warn('服务状态变更', payload.message)
			}
			loadData()
		})
	}

	return {
		services,
		loading,
		busy,
		loadData,
		startService,
		stopService,
		restartService,
		startConfiguredService,
		stopConfiguredService,
		markMain,
		toggleAutostart,
		browseFolder,
		bindEvents,
	}
}
