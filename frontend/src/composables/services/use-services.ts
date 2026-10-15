import { computed, ref } from 'vue'
import { listSoftware, listenEvent, startSoftware, stopSoftware } from '@/api'
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
	async function loadData(): Promise<void> {
		loading.value = true
		try {
			items.value = await listSoftware()
			const current = items.value.find((item) => item.id === busy.value)
			if (current && operation.value === 'start' && current.running) {
				clearBusy()
			}
			if (current && operation.value === 'stop' && !current.running) {
				clearBusy()
			}
		} catch (err: unknown) {
			toast.error('加载服务失败', formatError(err))
		} finally {
			loading.value = false
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
			if (payload.id === busy.value) {
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
		bindEvents,
	}
}
