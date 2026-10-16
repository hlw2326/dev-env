import { ref, computed } from 'vue'
import { listProjects, startSoftware, stopSoftware, loadSettings, loadSystemMetrics } from '@/api'
import { useServiceStatus } from '@/composables/services/use-service-status'
import { toast, formatError } from '@/utils'
import type { Project, ProjectTraffic, SystemMetrics } from '@/types'

// 项目列表状态
const projects = ref<Project[]>([])

// 已安装软件列表状态
const { items: installed, ready: servicesReady, checking, error: serviceError, refresh: refreshServices } = useServiceStatus()
const initialized = ref(false)
const dashboardError = ref('')

// 网关监听状态
const listening = ref(false)

// 网关监听端口
const nginxPort = ref(80)

// 主机资源指标
const systemMetrics = ref<SystemMetrics>({
	cpuPercent: 0,
	memoryTotalBytes: 0,
	memoryUsedBytes: 0,
	memoryPercent: 0,
})

// 数据加载中状态
const loading = ref(false)

// 正在操作的服务标识
const busyId = ref(0)

const SERVICE_REFRESH_TIMEOUT = 10_000
const SERVICE_REFRESH_INTERVAL = 250

// 概览页数据与操作组合式函数
export function useDashboard() {
	// 运行中项目数量
	const runningProjects = computed(() => {
		return projects.value.filter((item) => item.state === 'running').length
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

	// 当前环境中所有项目的累计访问流量
	const totalTraffic = computed<ProjectTraffic>(() => {
		const result: ProjectTraffic = {
			requestCount: 0,
			responseBytes: 0,
			errorCount: 0,
			averageResponseMs: 0,
			lastAccess: '',
		}
		let responseTimeTotal = 0
		for (const item of projects.value) {
			result.requestCount += item.traffic.requestCount
			result.responseBytes += item.traffic.responseBytes
			result.errorCount += item.traffic.errorCount
			responseTimeTotal += item.traffic.averageResponseMs * item.traffic.requestCount
			if (item.traffic.lastAccess > result.lastAccess) {
				result.lastAccess = item.traffic.lastAccess
			}
		}
		if (result.requestCount > 0) {
			result.averageResponseMs = responseTimeTotal / result.requestCount
		}
		return result
	})

	// 加载概览页全部数据
	async function loadData(silent = false): Promise<void> {
		if (!silent) {
			loading.value = true
		}
		try {
			const [projectList, , settings, metrics] = await Promise.all([
				listProjects(),
				refreshServices(silent),
				loadSettings(),
				loadSystemMetrics(),
			])
			projects.value = projectList
			listening.value = settings.listening
			nginxPort.value = settings.port
			systemMetrics.value = metrics
			initialized.value = true
			dashboardError.value = ''
		} catch (err: unknown) {
			dashboardError.value = formatError(err)
			if (!silent) {
				toast.error(formatError(err))
			}
		} finally {
			if (!silent) {
				loading.value = false
			}
		}
	}

	async function waitForServiceState(id: number, running: boolean): Promise<void> {
		const deadline = Date.now() + SERVICE_REFRESH_TIMEOUT
		do {
			await loadData(true)
			const current = installed.value.find((item) => item.id === id)
			if (current && current.running === running) {
				return
			}
			if (Date.now() >= deadline) {
				return
			}
			await new Promise<void>((resolve) => setTimeout(resolve, SERVICE_REFRESH_INTERVAL))
		} while (Date.now() < deadline)
	}

	// 启动指定守护进程服务
	async function startService(id: number): Promise<void> {
		if (!servicesReady.value || checking.value || serviceError.value !== '') return
		busyId.value = id
		try {
			const result = await startSoftware(id)
			if (!result.ok) {
				toast.error('启动失败', result.message)
				return
			}
			toast.success('服务已启动')
			await waitForServiceState(id, true)
		} catch (err: unknown) {
			toast.error(formatError(err))
		} finally {
			busyId.value = 0
		}
	}

	// 停止指定守护进程服务
	async function stopService(id: number): Promise<void> {
		if (!servicesReady.value || checking.value || serviceError.value !== '') return
		busyId.value = id
		try {
			const result = await stopSoftware(id)
			if (!result.ok) {
				toast.error('停止失败', result.message)
				return
			}
			toast.info('服务已停止')
			await waitForServiceState(id, false)
		} catch (err: unknown) {
			toast.error(formatError(err))
		} finally {
			busyId.value = 0
		}
	}

	return {
		servicesReady,
		checking,
		serviceError,
		initialized,
		dashboardError,
		projects,
		installed,
		readySoftware,
		daemonList,
		listening,
		nginxPort,
		systemMetrics,
		loading,
		busyId,
		runningProjects,
		runningDaemons,
		totalTraffic,
		loadData,
		startService,
		stopService,
	}
}
