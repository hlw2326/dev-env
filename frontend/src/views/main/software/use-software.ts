import { ref, reactive, computed } from 'vue'
import {
	loadMarket,
	listSoftware,
	listVersions,
	installSoftware,
	cancelInstall,
	removeSoftware,
	setMain,
	startSoftware,
	stopSoftware,
	setAutostart,
	openSoftware,
	listenEvent,
} from '@/api'
import { toast, formatError } from '@/utils'
import type {
	MarketItem,
	Software,
	SoftwareVersion,
	ProgressEvent,
	CompletedEvent,
	ServiceEvent,
} from '@/types'

// 已安装软件按类型分组项
export interface SoftwareGroup {
	// 软件类型标识
	kind: string
	// 软件显示名称
	name: string
	// 该类型下的版本列表
	items: Software[]
}

// 当前激活顶层页签
const tab = ref('market')

// 当前分类筛选项
const category = ref('全部')

// 搜索关键词
const search = ref('')

// 软件商店目录列表
const marketItems = ref<MarketItem[]>([])

// 已安装软件版本列表
const installedList = ref<Software[]>([])

// 软件商店加载错误消息
const marketError = ref('')

// 实时安装进度映射表
const progressMap = reactive<Record<string, { progress: number; message: string }>>({})

// 版本选择安装弹窗显隐状态
const modalVisible = ref(false)

// 当前准备安装的软件项
const marketItem = ref<MarketItem | null>(null)

// 可选安装版本列表
const versionOptions = ref<SoftwareVersion[]>([])

// 版本列表加载状态
const versionsLoading = ref(false)

// 版本列表加载错误消息
const versionsError = ref('')

// 软件设置弹窗显隐状态
const configVisible = ref(false)

// 当前正在配置的已安装软件
const activeSoftware = ref<Software | null>(null)

// 服务启停中的按钮状态
const serviceBusy = reactive<Record<number, boolean>>({})

// 服务当前操作方向
const serviceOperation = reactive<Record<number, string>>({})

// 软件商店与已安装管理组合式函数
export function useSoftware() {
	// 筛选后的软件商店列表
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

	// 筛选后的已安装软件列表
	const filteredInstalled = computed(() => {
		const keyword = search.value.trim().toLowerCase()
		return installedList.value.filter((item) => {
			if (keyword === '') {
				return true
			}
			if (item.kind.toLowerCase().includes(keyword)) {
				return true
			}
			if (item.name.toLowerCase().includes(keyword)) {
				return true
			}
			return item.version.toLowerCase().includes(keyword)
		})
	})

	// 按软件类型分组的已安装列表
	const installedGroups = computed<SoftwareGroup[]>(() => {
		const map = new Map<string, SoftwareGroup>()
		for (const item of filteredInstalled.value) {
			const existing = map.get(item.kind)
			if (existing) {
				existing.items.push(item)
			} else {
				map.set(item.kind, {
					kind: item.kind,
					name: item.name,
					items: [item],
				})
			}
		}
		return Array.from(map.values())
	})

	// 已成功安装的软件数量
	const installedCount = computed(() => {
		return installedList.value.filter((item) => item.state === 'installed').length
	})

	// 刷新商店与已安装列表
	async function refreshAll(): Promise<void> {
		marketError.value = ''
		try {
			const [market, installed] = await Promise.all([loadMarket(), listSoftware()])
			marketItems.value = market
			installedList.value = installed
			for (const item of installed) {
				const operation = serviceOperation[item.id]
				if (operation === 'start' && item.running) {
					clearServiceBusy(item.id)
				}
				if (operation === 'stop' && !item.running) {
					clearServiceBusy(item.id)
				}
			}
			if (activeSoftware.value !== null) {
				const activeId = activeSoftware.value.id
				const found = installed.find((item) => item.id === activeId)
				if (found) {
					activeSoftware.value = found
				}
			}
		} catch (err: unknown) {
			marketError.value = formatError(err)
		}
	}

	// 打开指定软件的版本选择弹窗
	async function openInstall(item: MarketItem): Promise<void> {
		marketItem.value = item
		modalVisible.value = true
		await fetchVersions(item.kind)
	}

	// 拉取指定软件类型的可安装版本清单
	async function fetchVersions(kind: string): Promise<void> {
		versionsLoading.value = true
		versionsError.value = ''
		versionOptions.value = []
		try {
			versionOptions.value = await listVersions(kind)
			marketItems.value = await loadMarket()
		} catch (err: unknown) {
			versionsError.value = formatError(err)
		} finally {
			versionsLoading.value = false
		}
	}

	// 发起软件版本安装任务
	async function startInstall(kind: string, version: string): Promise<void> {
		try {
			const result = await installSoftware({ kind, version })
			if (!result.ok) {
				toast.error('安装失败', result.message, 3500)
				return
			}
			const key = `${kind}:${version}`
			progressMap[key] = { progress: 1, message: '准备下载...' }
			await refreshAll()
		} catch (err: unknown) {
			toast.error('安装失败', formatError(err))
		}
	}

	// 取消正在进行的安装任务
	async function stopInstall(kind: string, version: string): Promise<void> {
		try {
			const result = await cancelInstall(kind, version)
			if (!result.ok) {
				toast.error('取消失败', result.message)
				return
			}
			const key = `${kind}:${version}`
			delete progressMap[key]
			await refreshAll()
		} catch (err: unknown) {
			toast.error('取消失败', formatError(err))
		}
	}

	// 卸载指定软件版本
	async function removeItem(item: Software): Promise<void> {
		try {
			const result = await removeSoftware(item.id)
			if (!result.ok) {
				toast.error('卸载被拒绝', result.message, 4000)
				return
			}
			if (activeSoftware.value !== null && activeSoftware.value.id === item.id) {
				configVisible.value = false
				activeSoftware.value = null
			}
			toast.info('已卸载', `${item.kind} ${item.version}`)
			await refreshAll()
		} catch (err: unknown) {
			toast.error('卸载失败', formatError(err))
		}
	}

	// 设为该软件类型的默认版本
	async function markMain(item: Software): Promise<void> {
		try {
			const result = await setMain(item.id)
			if (!result.ok) {
				toast.error('设置默认失败', result.message)
				return
			}
			toast.success('已设为默认版本', `${item.kind} ${item.version}`)
			await refreshAll()
		} catch (err: unknown) {
			toast.error('设置默认失败', formatError(err))
		}
	}

	// 启动守护进程服务
	async function startItem(item: Software): Promise<void> {
		setServiceBusy(item.id, 'start')
		try {
			const result = await startSoftware(item.id)
			if (!result.ok) {
				clearServiceBusy(item.id)
				toast.error('启动失败', result.message, 4000)
				return
			}
			toast.info('服务启动中', `${item.kind} ${item.version}`)
			await refreshAll()
		} catch (err: unknown) {
			clearServiceBusy(item.id)
			toast.error('启动失败', formatError(err))
		}
	}

	// 停止守护进程服务
	async function stopItem(item: Software): Promise<void> {
		setServiceBusy(item.id, 'stop')
		try {
			const result = await stopSoftware(item.id)
			if (!result.ok) {
				clearServiceBusy(item.id)
				toast.error('停止失败', result.message, 3500)
				return
			}
			toast.info('服务停止中', `${item.kind} ${item.version}`)
			await refreshAll()
		} catch (err: unknown) {
			clearServiceBusy(item.id)
			toast.error('停止失败', formatError(err))
		}
	}

	// 设置服务操作中的状态
	function setServiceBusy(id: number, operation: string): void {
		serviceBusy[id] = true
		serviceOperation[id] = operation
	}

	// 清理服务操作中的状态
	function clearServiceBusy(id: number): void {
		delete serviceBusy[id]
		delete serviceOperation[id]
	}

	// 查询指定服务是否正在操作
	function isServiceBusy(id: number): boolean {
		return serviceBusy[id] === true
	}

	// 切换守护进程随应用启动开关
	async function toggleAutostart(item: Software, on: boolean): Promise<void> {
		try {
			const result = await setAutostart(item.id, on)
			if (!result.ok) {
				toast.error('设置失败', result.message)
				return
			}
			await refreshAll()
		} catch (err: unknown) {
			toast.error('设置失败', formatError(err))
		}
	}

	// 打开软件安装或数据目录
	async function browseDir(id: number): Promise<void> {
		try {
			const result = await openSoftware(id)
			if (!result.ok) {
				toast.error('打开目录失败', result.message)
			}
		} catch (err: unknown) {
			toast.error('打开目录失败', formatError(err))
		}
	}

	// 打开软件配置弹窗
	function openConfig(item: Software): void {
		activeSoftware.value = item
		configVisible.value = true
	}

	// 按软件类型打开已安装版本配置
	function showConfig(kind: string): void {
		const found = installedList.value.find((item) => item.kind === kind && item.state === 'installed')
		if (found) {
			openConfig(found)
		} else {
			tab.value = 'installed'
		}
	}

	// 绑定安装进度与服务状态事件监听
	function bindEvents(): () => void {
		const stops: Array<() => void> = []
		stops.push(
			listenEvent('task:progress', (payload: ProgressEvent) => {
				const key = `${payload.kind}:${payload.version}`
				progressMap[key] = {
					progress: payload.progress,
					message: payload.message,
				}
			}),
		)
		stops.push(
			listenEvent('task:completed', (payload: CompletedEvent) => {
				const key = `${payload.kind}:${payload.version}`
				delete progressMap[key]
				if (payload.ok) {
					toast.success('安装完成', `${payload.kind} ${payload.version}`)
					if (marketItem.value && marketItem.value.kind === payload.kind) {
						fetchVersions(payload.kind)
					}
				} else {
					toast.error('安装失败', payload.error, 4500)
				}
				refreshAll()
			}),
		)
		stops.push(
			listenEvent('service:status-changed', (payload: ServiceEvent) => {
				if (payload.id > 0) {
					clearServiceBusy(payload.id)
				}
				if (payload.message !== '') {
					toast.warn('服务状态变更', payload.message, 4000)
				}
				refreshAll()
			}),
		)
		return () => {
			for (const stop of stops) {
				stop()
			}
		}
	}

	return {
		tab,
		category,
		search,
		marketItems,
		filteredMarket,
		installedList,
		filteredInstalled,
		installedGroups,
		installedCount,
		marketError,
		progressMap,
		modalVisible,
		marketItem,
		versionOptions,
		versionsLoading,
		versionsError,
		configVisible,
		activeSoftware,
		refreshAll,
		openInstall,
		fetchVersions,
		startInstall,
		stopInstall,
		removeItem,
		markMain,
		startItem,
		stopItem,
		isServiceBusy,
		toggleAutostart,
		browseDir,
		openConfig,
		showConfig,
		bindEvents,
	}
}
