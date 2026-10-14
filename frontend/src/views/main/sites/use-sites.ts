import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { EventsOn } from '../../../../wailsjs/runtime/runtime'
import {
	listSites,
	createSite,
	updateSite,
	startSite,
	stopSite,
	restartSite,
	deleteSite,
	openURL,
	openFolder,
	loadSettings,
	listSoftware,
	listVersions,
	installSoftware,
} from '@/api'
import type {
	Site,
	CreateSiteInput,
	UpdateSiteInput,
	Software,
	SiteStatusEvent,
	TaskCompletedEvent,
} from '@/types'

export function useSites() {
	const toast = useToast()
	const sites = ref<Site[]>([])
	const softwareList = ref<Software[]>([])
	const defaultRoot = ref('')
	const loading = ref(false)
	const search = ref('')
	const kindFilter = ref('all')
	const createVisible = ref(false)
	const activeSite = ref<Site | null>(null)
	const detailVisible = ref(false)
	const installingKind = ref('')

	const filteredSites = computed(() => {
		const keyword = search.value.trim().toLowerCase()
		return sites.value.filter((item) => {
			if (kindFilter.value !== 'all' && item.kind !== kindFilter.value) {
				return false
			}
			if (keyword === '') {
				return true
			}
			if (item.name.toLowerCase().includes(keyword)) {
				return true
			}
			return item.hosts.some((h) => h.toLowerCase().includes(keyword))
		})
	})

	async function refreshAll() {
		loading.value = true
		try {
			const [siteItems, softItems, cfg] = await Promise.all([
				listSites(),
				listSoftware(),
				loadSettings(),
			])
			sites.value = siteItems
			softwareList.value = softItems
			defaultRoot.value = cfg.root
			if (activeSite.value !== null) {
				const curId = activeSite.value.id
				const found = siteItems.find((s) => s.id === curId)
				if (found) {
					activeSite.value = found
				}
			}
		} catch (err) {
			toast.add({ severity: 'error', summary: '加载失败', detail: String(err), life: 3000 })
		} finally {
			loading.value = false
		}
	}

	async function handleCreate(input: CreateSiteInput): Promise<boolean> {
		try {
			await createSite(input)
			toast.add({ severity: 'success', summary: '创建成功', detail: `网站 ${input.name} 已创建`, life: 2500 })
			await refreshAll()
			return true
		} catch (err) {
			toast.add({ severity: 'error', summary: '创建失败', detail: String(err), life: 3500 })
			return false
		}
	}

	async function handleUpdate(input: UpdateSiteInput): Promise<boolean> {
		try {
			const updated = await updateSite(input)
			activeSite.value = updated
			await refreshAll()
			toast.add({ severity: 'success', summary: '已保存', detail: '站点设置已生效', life: 2000 })
			return true
		} catch (err) {
			toast.add({ severity: 'error', summary: '保存失败', detail: String(err), life: 3500 })
			return false
		}
	}

	async function handleToggleState(item: Site) {
		if (item.state === 'running') {
			const res = await stopSite(item.id)
			if (!res.ok) {
				toast.add({ severity: 'error', summary: '停止失败', detail: res.message, life: 3500 })
				return
			}
			toast.add({ severity: 'info', summary: '已停止', detail: item.name, life: 2000 })
		} else {
			const res = await startSite(item.id)
			if (!res.ok) {
				toast.add({ severity: 'error', summary: '启动失败', detail: res.message, life: 3500 })
				return
			}
			toast.add({ severity: 'success', summary: '已启动', detail: item.name, life: 2000 })
		}
		await refreshAll()
	}

	async function handleRestart(item: Site) {
		const res = await restartSite(item.id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '重启失败', detail: res.message, life: 3500 })
			return
		}
		toast.add({ severity: 'success', summary: '已重启', detail: item.name, life: 2000 })
		await refreshAll()
	}

	async function handleDelete(item: Site) {
		const res = await deleteSite(item.id)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '删除失败', detail: res.message, life: 3500 })
			return
		}
		if (activeSite.value !== null && activeSite.value.id === item.id) {
			detailVisible.value = false
			activeSite.value = null
		}
		toast.add({ severity: 'info', summary: '已删除', detail: res.message, life: 2000 })
		await refreshAll()
	}

	async function handleQuickInstall(kind: string) {
		if (installingKind.value !== '') {
			return
		}
		installingKind.value = kind
		try {
			const versions = await listVersions(kind)
			if (versions.length === 0) {
				toast.add({ severity: 'error', summary: '安装失败', detail: '未获取到可用版本', life: 3500 })
				installingKind.value = ''
				return
			}
			let target = versions[0].version
			for (const v of versions) {
				if (v.recommended) {
					target = v.version
					break
				}
			}
			const res = await installSoftware({ kind, version: target })
			if (!res.ok) {
				toast.add({ severity: 'error', summary: '安装失败', detail: res.message, life: 3500 })
				installingKind.value = ''
				return
			}
			toast.add({ severity: 'info', summary: '正在安装', detail: `正在后台下载安装 ${kind} ${target}`, life: 3000 })
		} catch (err) {
			toast.add({ severity: 'error', summary: '拉取版本失败', detail: String(err), life: 3500 })
			installingKind.value = ''
		}
	}

	function handleOpenSite(item: Site) {
		if (item.url !== '') {
			openURL(item.url)
		}
	}

	async function handleOpenFolder(path: string) {
		const res = await openFolder(path)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '打开目录失败', detail: res.message, life: 3000 })
		}
	}

	function openDetail(item: Site) {
		activeSite.value = item
		detailVisible.value = true
	}

	const stops: Array<() => void> = []

	onMounted(() => {
		refreshAll()
		stops.push(
			EventsOn('site:status-changed', (_payload: SiteStatusEvent) => {
				refreshAll()
			})
		)
		stops.push(
			EventsOn('task:completed', (payload: TaskCompletedEvent) => {
				if (installingKind.value === payload.kind) {
					installingKind.value = ''
				}
				if (payload.ok) {
					toast.add({ severity: 'success', summary: '环境安装完成', detail: `${payload.kind} ${payload.version} 已就绪`, life: 3000 })
				} else {
					toast.add({ severity: 'error', summary: '环境安装失败', detail: payload.error, life: 4000 })
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
		sites,
		filteredSites,
		softwareList,
		defaultRoot,
		loading,
		search,
		kindFilter,
		createVisible,
		activeSite,
		detailVisible,
		installingKind,
		refreshAll,
		handleCreate,
		handleUpdate,
		handleToggleState,
		handleRestart,
		handleDelete,
		handleQuickInstall,
		handleOpenSite,
		handleOpenFolder,
		openDetail,
	}
}
