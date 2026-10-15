import { ref, computed } from 'vue'
import { listSites, listSoftware, startSoftware, stopSoftware, loadSettings } from '@/api'
import { toast, formatError } from '@/utils'
import type { Site, Software } from '@/types'

// 站点列表状态
const sites = ref<Site[]>([])

// 已安装软件列表状态
const installed = ref<Software[]>([])

// 网关监听状态
const listening = ref(false)

// 网关监听端口
const nginxPort = ref(80)

// 数据加载中状态
const loading = ref(false)

// 正在操作的服务标识
const busyId = ref(0)

// 概览页数据与操作组合式函数
export function useDashboard() {
	// 运行中站点数量
	const runningSites = computed(() => {
		return sites.value.filter((item) => item.state === 'running').length
	})

	// 已安装就绪的软件列表
	const readySoftware = computed(() => {
		return installed.value.filter((item) => item.state === 'installed')
	})

	// 守护进程类软件列表
	const daemonList = computed(() => {
		return readySoftware.value.filter((item) => item.service)
	})

	// 运行中守护进程数量
	const runningDaemons = computed(() => {
		return daemonList.value.filter((item) => item.running).length
	})

	// 加载概览页全部数据
	async function loadData(): Promise<void> {
		loading.value = true
		try {
			const [siteList, softwareList, settings] = await Promise.all([
				listSites(),
				listSoftware(),
				loadSettings(),
			])
			sites.value = siteList
			installed.value = softwareList
			listening.value = settings.listening
			nginxPort.value = settings.port
		} catch (err: unknown) {
			toast.error(formatError(err))
		} finally {
			loading.value = false
		}
	}

	// 启动指定守护进程服务
	async function startService(id: number): Promise<void> {
		busyId.value = id
		try {
			const result = await startSoftware(id)
			if (!result.ok) {
				toast.error('启动失败', result.message)
				return
			}
			toast.success('服务已启动')
			await loadData()
		} catch (err: unknown) {
			toast.error(formatError(err))
		} finally {
			busyId.value = 0
		}
	}

	// 停止指定守护进程服务
	async function stopService(id: number): Promise<void> {
		busyId.value = id
		try {
			const result = await stopSoftware(id)
			if (!result.ok) {
				toast.error('停止失败', result.message)
				return
			}
			toast.info('服务已停止')
			await loadData()
		} catch (err: unknown) {
			toast.error(formatError(err))
		} finally {
			busyId.value = 0
		}
	}

	return {
		sites,
		installed,
		readySoftware,
		daemonList,
		listening,
		nginxPort,
		loading,
		busyId,
		runningSites,
		runningDaemons,
		loadData,
		startService,
		stopService,
	}
}
