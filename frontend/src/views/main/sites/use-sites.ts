import { ref, computed } from 'vue'
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
	listProjectPlugins,
	runProjectPlugin,
	listVersions,
	installSoftware,
	listenEvent,
} from '@/api'
import { toast, formatError } from '@/utils'
import type {
	Site,
	CreateInput,
	UpdateInput,
	Software,
	SiteEvent,
	CompletedEvent,
	ProjectPlugin,
	PluginActionInput,
	ProjectPluginEvent,
} from '@/types'

// 网站列表状态
const sites = ref<Site[]>([])

// 已安装软件列表状态
const softwareList = ref<Software[]>([])

// 默认站点父目录
const defaultRoot = ref('')

// 加载中状态
const loading = ref(false)

// 搜索关键词
const search = ref('')

// 站点类型筛选值
const kindFilter = ref('all')

// 创建弹窗显隐状态
const createVisible = ref(false)

// 当前查看详情的站点
const activeSite = ref<Site | null>(null)

// 详情弹窗显隐状态
const detailVisible = ref(false)

// 正在一键安装的软件类型
const installingKind = ref('')

// 当前项目可用插件
const projectPlugins = ref<ProjectPlugin[]>([])

// 当前正在执行的插件
const pluginBusy = ref('')

// 站点管理组合式函数
export function useSites() {
	// 筛选后的站点列表
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
			return item.hosts.some((host) => host.toLowerCase().includes(keyword))
		})
	})

	// 刷新站点与环境列表
	async function refreshAll(): Promise<void> {
		loading.value = true
		try {
			const [siteItems, softwareItems, settings] = await Promise.all([
				listSites(),
				listSoftware(),
				loadSettings(),
			])
			sites.value = siteItems
			softwareList.value = softwareItems
			defaultRoot.value = settings.root
			if (activeSite.value !== null) {
				const activeId = activeSite.value.id
				const found = siteItems.find((item) => item.id === activeId)
				if (found) {
					activeSite.value = found
				}
			}
		} catch (err: unknown) {
			toast.error('加载失败', formatError(err))
		} finally {
			loading.value = false
		}
	}

	// 加载项目详情插件
	async function loadPlugins(id: number): Promise<void> {
		projectPlugins.value = []
		pluginBusy.value = ''
		try {
			projectPlugins.value = await listProjectPlugins(id)
		} catch (err: unknown) {
			toast.error('加载项目插件失败', formatError(err))
		}
	}

	// 提交创建新网站
	async function submitCreate(input: CreateInput): Promise<boolean> {
		try {
			await createSite(input)
			toast.success('创建成功', `项目 ${input.name} 已创建`)
			await refreshAll()
			return true
		} catch (err: unknown) {
			toast.error('创建失败', formatError(err), 3500)
			return false
		}
	}

	// 提交更新站点配置
	async function submitUpdate(input: UpdateInput): Promise<boolean> {
		try {
			const updated = await updateSite(input)
			activeSite.value = updated
			await refreshAll()
			toast.success('已保存', '项目设置已生效')
			return true
		} catch (err: unknown) {
			toast.error('保存失败', formatError(err), 3500)
			return false
		}
	}

	// 切换站点启动或停止状态
	async function toggleSite(item: Site): Promise<void> {
		try {
			if (item.state === 'running') {
				const result = await stopSite(item.id)
				if (!result.ok) {
					toast.error('停止失败', result.message, 3500)
					return
				}
				toast.info('已停止', item.name)
			} else {
				const result = await startSite(item.id)
				if (!result.ok) {
					toast.error('启动失败', result.message, 3500)
					return
				}
				toast.success('已启动', item.name)
			}
			await refreshAll()
		} catch (err: unknown) {
			toast.error('操作失败', formatError(err))
		}
	}

	// 重启指定站点
	async function restartItem(item: Site): Promise<void> {
		try {
			const result = await restartSite(item.id)
			if (!result.ok) {
				toast.error('重启失败', result.message, 3500)
				return
			}
			toast.success('已重启', item.name)
			await refreshAll()
		} catch (err: unknown) {
			toast.error('重启失败', formatError(err))
		}
	}

	// 删除指定站点记录
	async function removeSite(item: Site): Promise<void> {
		try {
			const result = await deleteSite(item.id)
			if (!result.ok) {
				toast.error('删除失败', result.message, 3500)
				return
			}
			if (activeSite.value !== null && activeSite.value.id === item.id) {
				detailVisible.value = false
				activeSite.value = null
			}
			toast.info('已删除', result.message)
			await refreshAll()
		} catch (err: unknown) {
			toast.error('删除失败', formatError(err))
		}
	}

	// 一键安装缺失的推荐版本运行环境
	async function quickInstall(kind: string): Promise<void> {
		if (installingKind.value !== '') {
			return
		}
		installingKind.value = kind
		try {
			const versions = await listVersions(kind)
			if (versions.length === 0) {
				toast.error('安装失败', '未获取到可用版本', 3500)
				installingKind.value = ''
				return
			}
			let target = versions[0].version
			for (const item of versions) {
				if (item.recommended) {
					target = item.version
					break
				}
			}
			const result = await installSoftware({ kind, version: target })
			if (!result.ok) {
				toast.error('安装失败', result.message, 3500)
				installingKind.value = ''
				return
			}
			toast.info('正在安装', `正在后台下载安装 ${kind} ${target}`)
		} catch (err: unknown) {
			toast.error('拉取版本失败', formatError(err), 3500)
			installingKind.value = ''
		}
	}

	// 在默认浏览器中打开站点
	function openSite(item: Site): void {
		if (item.url !== '') {
			openURL(item.url).catch((err: unknown) => {
				toast.error('打开链接失败', formatError(err))
			})
		}
	}

	// 在资源管理器中打开站点目录
	async function browseFolder(path: string): Promise<void> {
		try {
			const result = await openFolder(path)
			if (!result.ok) {
				toast.error('打开目录失败', result.message)
			}
		} catch (err: unknown) {
			toast.error('打开目录失败', formatError(err))
		}
	}

	// 打开站点详情弹窗
	function openDetail(item: Site): void {
		activeSite.value = item
		detailVisible.value = true
		loadPlugins(item.id)
	}

	// 执行项目插件操作
	async function runPlugin(input: PluginActionInput): Promise<void> {
		if (pluginBusy.value !== '') {
			return
		}
		pluginBusy.value = input.pluginId
		try {
			const result = await runProjectPlugin(input)
			if (!result.ok) {
				pluginBusy.value = ''
				toast.error('插件操作失败', result.message, 4000)
				return
			}
			toast.info('插件任务已启动', input.pluginId)
		} catch (err: unknown) {
			pluginBusy.value = ''
			toast.error('插件操作失败', formatError(err))
		}
	}

	// 绑定站点与安装完成事件监听
	function bindEvents(): () => void {
		const stops: Array<() => void> = []
		stops.push(
			listenEvent('site:status-changed', (payload: SiteEvent) => {
				if (payload.id > 0) {
					refreshAll()
				}
			}),
		)
		stops.push(
			listenEvent('task:completed', (payload: CompletedEvent) => {
				if (installingKind.value === payload.kind) {
					installingKind.value = ''
				}
				if (payload.ok) {
					toast.success('环境安装完成', `${payload.kind} ${payload.version} 已就绪`)
				} else {
					toast.error('环境安装失败', payload.error, 4000)
				}
				refreshAll()
				const current = activeSite.value
				if (current !== null) {
					loadPlugins(current.id)
				}
			}),
		)
		stops.push(
			listenEvent('project:plugin-status-changed', (payload: ProjectPluginEvent) => {
				const current = activeSite.value
				if (current !== null && payload.id === current.id && !payload.running) {
					pluginBusy.value = ''
					if (payload.ok) {
						toast.success('插件操作完成', payload.plugin)
					} else {
						toast.error('插件操作失败', payload.message, 4500)
					}
				}
			})
		)
		return () => {
			for (const stop of stops) {
				stop()
			}
		}
	}

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
	projectPlugins,
	pluginBusy,
		refreshAll,
		submitCreate,
		submitUpdate,
		toggleSite,
		restartItem,
		removeSite,
		quickInstall,
		openSite,
		browseFolder,
		openDetail,
		loadPlugins,
		runPlugin,
		bindEvents,
	}
}
