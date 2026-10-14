import { ref, computed } from 'vue'
import { listSites, listSoftware, startSoftware, stopSoftware, loadSettings } from '@/api'
import { toast } from '@/utils'
import type { Site, Software } from '@/types'

const sites = ref<Site[]>([])
const installed = ref<Software[]>([])
const listening = ref(false)
const gatewayPort = ref(80)
const loading = ref(false)
const busyId = ref(0)

export function useDashboard() {
	const runningSites = computed(() => {
		return sites.value.filter((item) => item.state === 'running').length
	})

	const readySoftware = computed(() => {
		return installed.value.filter((item) => item.state === 'installed')
	})

	const daemonList = computed(() => {
		return readySoftware.value.filter(
			(item) => item.kind === 'mysql' || item.kind === 'redis' || item.kind === 'nginx' || item.kind === 'apache'
		)
	})

	const runningDaemons = computed(() => {
		return daemonList.value.filter((item) => item.running).length
	})

	async function loadData() {
		loading.value = true
		try {
			const [siteList, softList, sysInfo] = await Promise.all([
				listSites(),
				listSoftware(),
				loadSettings(),
			])
			sites.value = siteList
			installed.value = softList
			listening.value = sysInfo.listening
			gatewayPort.value = sysInfo.port
		} catch (err) {
			toast.error(String(err))
		} finally {
			loading.value = false
		}
	}

	async function handleStart(id: number) {
		busyId.value = id
		try {
			await startSoftware(id)
			toast.success('服务已启动')
			await loadData()
		} catch (err) {
			toast.error(String(err))
		} finally {
			busyId.value = 0
		}
	}

	async function handleStop(id: number) {
		busyId.value = id
		try {
			await stopSoftware(id)
			toast.info('服务已停止')
			await loadData()
		} catch (err) {
			toast.error(String(err))
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
		gatewayPort,
		loading,
		busyId,
		runningSites,
		runningDaemons,
		loadData,
		handleStart,
		handleStop,
	}
}
