import { ref, computed } from 'vue'
import { listSites, listAllRuntimes, listServices, startService, stopService } from '@/api'
import type { Site, RuntimeVersion, ServiceInstance } from '@/types'

// 网站列表状态
const sites = ref<Site[]>([])
// 运行环境列表状态
const runtimes = ref<RuntimeVersion[]>([])
// 服务实例列表状态
const services = ref<ServiceInstance[]>([])
// 加载状态
const loading = ref(false)
// 正在操作的服务编号
const busyId = ref<number | null>(null)

// 控制台状态钩子
export function useDashboard() {
	// 运行中网站数量
	const runningSites = computed(() => sites.value.filter((s) => s.status === 'running').length)
	// 运行中服务数量
	const runningServices = computed(() => services.value.filter((s) => s.status === 'running').length)

	// 加载概览数据
	async function loadData() {
		loading.value = true
		try {
			const [st, rt, srv] = await Promise.all([
				listSites(),
				listAllRuntimes(),
				listServices(),
			])
			sites.value = st
			runtimes.value = rt
			services.value = srv
		} finally {
			loading.value = false
		}
	}

	// 启动服务实例
	async function startSrv(id: number) {
		busyId.value = id
		try {
			await startService(id)
			toast.success('服务启动成功')
			await loadData()
		} catch (err) {
			toast.error(String(err))
		} finally {
			busyId.value = null
		}
	}

	// 停止服务实例
	async function stopSrv(id: number) {
		busyId.value = id
		try {
			await stopService(id)
			toast.info('服务已成功停止')
			await loadData()
		} catch (err) {
			toast.error(String(err))
		} finally {
			busyId.value = null
		}
	}

	return {
		sites,
		runtimes,
		services,
		loading,
		busyId,
		runningSites,
		runningServices,
		loadData,
		startSrv,
		stopSrv,
	}
}
