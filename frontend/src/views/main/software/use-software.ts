import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { EventsOn } from '../../../../wailsjs/runtime/runtime'
import {
	loadMarket,
	listSoftware,
	listVersions,
	installSoftware,
	cancelInstall,
	removeSoftware,
	setMainSoftware,
	startSoftware,
	stopSoftware,
	setSoftwareAutostart,
	openSoftwareFolder,
} from '@/api'
import type {
	MarketItem,
	Software,
	SoftwareVersion,
	TaskProgressEvent,
	TaskCompletedEvent,
	ServiceStatusEvent,
} from '@/types'

export function useSoftware() {
	const toast = useToast()
	const tab = ref('market')
	const category = ref('全部')
	const search = ref('')
	const marketItems = ref<MarketItem[]>([])
	const installedList = ref<Software[]>([])
	const marketError = ref('')

	const progressMap = reactive<Record<string, { progress: number; message: string }>>({})
	const installModalVisible = ref(false)
	const activeMarketItem = ref<MarketItem | null>(null)
	const versionOptions = ref<SoftwareVersion[]>([])
	const versionsLoading = ref(false)
	const versionsError = ref('')

	const configVisible = ref(false)
	const activeSoftware = ref<Software | null>(null)

	const filteredMarket = computed(() => {
		const keyword = search.value.trim().toLowerCase()
		return marketItems.value.filter((item) => {
			if (category.value !== '全部' && item.category !== category.value) {
				return false
			}
			if (keyword === '') {
				return true
			}
			if (item.name.toLowerCase().includes(keyword)) {
				return true
			}
			return item.description.toLowerCase().includes(keyword)
		})
	})

	const filteredInstalled = computed(() => {
		const keyword = search.value.trim().toLowerCase()
		return installedList.value.filter((item) => {
			if (keyword === '') {
				return true
			}
			if (item.kind.toLowerCase().includes(keyword)) {
				return true
			}
			return item.version.toLowerCase().includes(keyword)
		})
	})

	const installedCount = computed(() => {
		return installedList.value.filter((item) => item.state === 'installed').length
	})

	async function refreshAll() {
		marketError.value = ''
		try {
			const [market, installed] = await Promise.all([loadMarket(), listSoftware()])
			marketItems.value = market
			installedList.value = installed
			if (activeSoftware.value !== null) {
				const curId = activeSoftware.value.id
				const found = installed.find((s) => s.id === curId)
				if (found) {
					activeSoftware.value = found
				}
			}
		} catch (err) {
			marketError.value = String(err)
		}
	}

	async function openInstallModal(item: MarketItem) {
		activeMarketItem.value = item
		installModalVisible.value = true
		await fetchVersionsFor(item.kind)
	}

	async function fetchVersionsFor(kind: string) {
		versionsLoading.value = true
		versionsError.value = ''
		versionOptions.value = []
		try {
			versionOptions.value = await listVersions(kind)
		} catch (err) {
			versionsError.value = String(err)
		} finally {
			versionsLoading.value = false
		}
	}

	async function handleInstall(kind: string, version: string) {
		const res = await installSoftware({ kind, version })
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '安装失败', detail: res.message, life: 3500 })
			return
		}
		const key = `${kind}:${version}`
		progressMap[key] = { progress: 1, message: '准备下载...' }
		await refreshAll()
	}

	async function handleCancelInstall(kind: string, version: string) {
		const res = await cancelInstall(kind, version)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '取消失败', detail: res.message, life: 3000 })
			return
		}
		const key = `${kind}:${version}`
		delete progressMap[key]
		await refreshAll()
	}

	async function handleRemove(item: Software) {
		const res = await removeSoftware(item.id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '卸载被拒绝', detail: res.message, life: 4000 })
			return
		}
		if (activeSoftware.value !== null && activeSoftware.value.id === item.id) {
			configVisible.value = false
			activeSoftware.value = null
		}
		toast.add({ severity: 'info', summary: '已卸载', detail: `${item.kind} ${item.version}`, life: 2500 })
		await refreshAll()
	}

	async function handleSetMain(item: Software) {
		const res = await setMainSoftware(item.id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '设置默认失败', detail: res.message, life: 3000 })
			return
		}
		toast.add({ severity: 'success', summary: '已设为默认版本', detail: `${item.kind} ${item.version}`, life: 2000 })
		await refreshAll()
	}

	async function handleStart(item: Software) {
		const res = await startSoftware(item.id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '启动失败', detail: res.message, life: 4000 })
			return
		}
		toast.add({ severity: 'success', summary: '服务已启动', detail: `${item.kind} ${item.version}`, life: 2000 })
		await refreshAll()
	}

	async function handleStop(item: Software) {
		const res = await stopSoftware(item.id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '停止失败', detail: res.message, life: 3500 })
			return
		}
		toast.add({ severity: 'info', summary: '服务已停止', detail: `${item.kind} ${item.version}`, life: 2000 })
		await refreshAll()
	}

	async function handleSetAutostart(item: Software, on: boolean) {
		const res = await setSoftwareAutostart(item.id, on)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '设置失败', detail: res.message, life: 3000 })
			return
		}
		await refreshAll()
	}

	async function handleOpenFolder(id: number) {
		const res = await openSoftwareFolder(id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '打开目录失败', detail: res.message, life: 3000 })
		}
	}

	function openConfigDrawer(item: Software) {
		activeSoftware.value = item
		configVisible.value = true
	}

	function openConfigByKind(kind: string) {
		const found = installedList.value.find((s) => s.kind === kind && s.state === 'installed')
		if (found) {
			openConfigDrawer(found)
		} else {
			tab.value = 'installed'
		}
	}

	const stops: Array<() => void> = []

	onMounted(() => {
		refreshAll()
		stops.push(
			EventsOn('task:progress', (payload: TaskProgressEvent) => {
				const key = `${payload.kind}:${payload.version}`
				progressMap[key] = {
					progress: payload.progress,
					message: payload.message,
				}
			})
		)
		stops.push(
			EventsOn('task:completed', (payload: TaskCompletedEvent) => {
				const key = `${payload.kind}:${payload.version}`
				delete progressMap[key]
				if (payload.ok) {
					toast.add({ severity: 'success', summary: '安装完成', detail: `${payload.kind} ${payload.version}`, life: 3000 })
					if (activeMarketItem.value && activeMarketItem.value.kind === payload.kind) {
						fetchVersionsFor(payload.kind)
					}
				} else {
					toast.add({ severity: 'error', summary: '安装失败', detail: payload.error, life: 4500 })
				}
				refreshAll()
			})
		)
		stops.push(
			EventsOn('service:status-changed', (payload: ServiceStatusEvent) => {
				if (payload.message !== '') {
					toast.add({ severity: 'warn', summary: '服务状态变更', detail: payload.message, life: 4000 })
				}
				refreshAll()
			})
		)
	})

	onUnmounted(() => {
		for (const stop of stops) {
			stop()
		}
	})

	return {
		tab,
		category,
		search,
		marketItems,
		filteredMarket,
		installedList,
		filteredInstalled,
		installedCount,
		marketError,
		progressMap,
		installModalVisible,
		activeMarketItem,
		versionOptions,
		versionsLoading,
		versionsError,
		configVisible,
		activeSoftware,
		refreshAll,
		openInstallModal,
		fetchVersionsFor,
		handleInstall,
		handleCancelInstall,
		handleRemove,
		handleSetMain,
		handleStart,
		handleStop,
		handleSetAutostart,
		handleOpenFolder,
		openConfigDrawer,
		openConfigByKind,
	}
}
