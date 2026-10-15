import { computed, ref } from 'vue'
import { listSoftware, listenEvent, stopSoftware } from '@/api'
import { formatError, toast } from '@/utils'
import type { ServiceEvent, Software } from '@/types'

// 已安装软件服务列表
const items = ref<Software[]>([])

// 页面加载状态
const loading = ref(false)

// 正在操作的服务标识
const busy = ref(0)

// 运行中服务页面组合式函数
export function useServices() {
	// 当前运行中的守护服务
	const running = computed(() => {
		return items.value.filter((item) => item.service && item.running)
	})

	// 加载服务运行状态
	async function loadData(): Promise<void> {
		loading.value = true
		try {
			items.value = await listSoftware()
		} catch (err: unknown) {
			toast.error('加载服务失败', formatError(err))
		} finally {
			loading.value = false
		}
	}

	// 停止指定运行服务
	async function stopService(id: number): Promise<void> {
		busy.value = id
		try {
			const result = await stopSoftware(id)
			if (!result.ok) {
				toast.error('停止失败', result.message)
				return
			}
			toast.info('服务已停止')
			await loadData()
		} catch (err: unknown) {
			toast.error('停止失败', formatError(err))
		} finally {
			busy.value = 0
		}
	}

	// 监听服务状态变化并刷新列表
	function bindEvents(): () => void {
		return listenEvent('service:status-changed', (payload: ServiceEvent) => {
			if (payload.message !== '') {
				toast.warn('服务状态变更', payload.message)
			}
			loadData()
		})
	}

	return {
		running,
		loading,
		busy,
		loadData,
		stopService,
		bindEvents,
	}
}
