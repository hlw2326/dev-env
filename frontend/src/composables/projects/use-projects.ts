import { ref, computed } from 'vue'
import {
	listProjects,
	createProject,
	updateProject,
	startProject,
	stopProject,
	restartProject,
	deleteProject,
	openURL,
	openFolder,
	loadSettings,
	listSoftware,
	listProjectPlugins,
	listProjectTemplates,
	runProjectPlugin,
	listVersions,
	installSoftware,
	listenEvent,
	listProjectGitStatuses,
} from '@/api'
import { toast, formatError } from '@/utils'
import type {
	Project,
	ProjectCreateInput,
	ProjectUpdateInput,
	Software,
	ProjectEvent,
	CompletedEvent,
	ProjectPlugin,
	ProjectPluginActionInput,
	ProjectPluginEvent,
	ProjectTemplate,
	ProjectGitStatus,
} from '@/types'

// 项目列表状态
const projects = ref<Project[]>([])

// 已安装软件列表状态
const softwareList = ref<Software[]>([])

// 后端注册的项目模板
const projectTemplates = ref<ProjectTemplate[]>([])

// 项目 Git 状态
const gitStatuses = ref<Record<number, ProjectGitStatus>>({})

// 默认项目父目录
const defaultRoot = ref('')

// 加载中状态
const loading = ref(false)

// 搜索关键词
const search = ref('')

// 项目类型筛选值
const kindFilter = ref('all')

// 创建弹窗显隐状态
const createVisible = ref(false)

// 当前查看详情的项目
const activeProject = ref<Project | null>(null)

// 详情弹窗显隐状态
const detailVisible = ref(false)

// 打开详情时默认显示的页签
const detailInitialTab = ref('overview')

// 正在一键安装的软件类型
const installingKind = ref('')

// 当前项目可用插件
const projectPlugins = ref<ProjectPlugin[]>([])

// 当前正在执行的插件
const pluginBusy = ref('')

let gitStatusRequest = 0

// 项目管理组合式函数
export function useProjects() {
	// 筛选后的项目列表
	const filteredProjects = computed(() => {
		const keyword = search.value.trim().toLowerCase()
		return projects.value.filter((item) => {
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

	// 刷新项目与环境列表
	async function refreshAll(): Promise<void> {
		loading.value = true
		try {
			const [projectItems, softwareItems, settings, templateItems] = await Promise.all([
				listProjects(),
				listSoftware(),
				loadSettings(),
				listProjectTemplates(),
			])
			projects.value = projectItems
			softwareList.value = softwareItems
			projectTemplates.value = templateItems
			defaultRoot.value = settings.root
			if (activeProject.value !== null) {
				const activeId = activeProject.value.id
				const found = projectItems.find((item) => item.id === activeId)
				if (found) {
					activeProject.value = found
				}
			}
			void refreshGitStatuses()
		} catch (err: unknown) {
			toast.error('加载失败', formatError(err))
		} finally {
			loading.value = false
		}
	}

	// 异步刷新所有项目 Git 状态，不阻塞项目列表首屏
	async function refreshGitStatuses(): Promise<void> {
		const request = ++gitStatusRequest
		try {
			const statuses = await listProjectGitStatuses()
			if (request !== gitStatusRequest) {
				return
			}
			const next: Record<number, ProjectGitStatus> = {}
			for (const status of statuses) {
				next[status.projectId] = status
			}
			gitStatuses.value = next
		} catch (err: unknown) {
			toast.warn('读取项目 Git 状态失败', formatError(err))
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

	// 提交创建新项目
	async function submitCreate(input: ProjectCreateInput): Promise<boolean> {
		try {
			await createProject(input)
			toast.success('创建成功', `项目 ${input.name} 已创建`)
			await refreshAll()
			return true
		} catch (err: unknown) {
			toast.error('创建失败', formatError(err), 3500)
			return false
		}
	}

	// 提交更新项目配置
	async function submitUpdate(input: ProjectUpdateInput): Promise<boolean> {
		try {
			const updated = await updateProject(input)
			activeProject.value = updated
			await refreshAll()
			toast.success('已保存', '项目设置已生效')
			return true
		} catch (err: unknown) {
			toast.error('保存失败', formatError(err), 3500)
			return false
		}
	}

	// 切换项目启动或停止状态
	async function toggleProject(item: Project): Promise<void> {
		try {
			if (item.state === 'running') {
				const result = await stopProject(item.id)
				if (!result.ok) {
					toast.error('停止失败', result.message, 3500)
					return
				}
				toast.info('已停止', item.name)
			} else {
				const result = await startProject(item.id)
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

	// 重启指定项目
	async function restartItem(item: Project): Promise<void> {
		try {
			const result = await restartProject(item.id)
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

	// 删除指定项目记录
	async function removeProject(item: Project): Promise<void> {
		try {
			const result = await deleteProject(item.id)
			if (!result.ok) {
				toast.error('删除失败', result.message, 3500)
				return
			}
			if (activeProject.value !== null && activeProject.value.id === item.id) {
				detailVisible.value = false
				activeProject.value = null
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

	// 在默认浏览器中打开项目
	function openProject(item: Project): void {
		if (item.url !== '') {
			openURL(item.url).catch((err: unknown) => {
				toast.error('打开链接失败', formatError(err))
			})
		}
	}

	// 在资源管理器中打开项目目录
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

	// 打开项目详情弹窗
	function openDetail(item: Project, tab = 'overview'): void {
		activeProject.value = item
		detailInitialTab.value = tab
		detailVisible.value = true
		loadPlugins(item.id)
	}

	function openGit(item: Project): void {
		openDetail(item, 'git')
	}

	// 执行项目插件操作
	async function runPlugin(input: ProjectPluginActionInput): Promise<void> {
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

	// 绑定项目与安装完成事件监听
	function bindEvents(): () => void {
		const stops: Array<() => void> = []
		stops.push(
			listenEvent('project:status-changed', (payload: ProjectEvent) => {
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
				const current = activeProject.value
				if (current !== null) {
					loadPlugins(current.id)
				}
			}),
		)
		stops.push(
			listenEvent('project:plugin-status-changed', (payload: ProjectPluginEvent) => {
				const current = activeProject.value
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
		projects,
		filteredProjects,
		softwareList,
		projectTemplates,
		gitStatuses,
		defaultRoot,
		loading,
		search,
		kindFilter,
		createVisible,
		activeProject,
		detailVisible,
		detailInitialTab,
	installingKind,
	projectPlugins,
	pluginBusy,
		refreshAll,
		refreshGitStatuses,
		submitCreate,
		submitUpdate,
		toggleProject,
		restartItem,
		removeProject,
		quickInstall,
		openProject,
		browseFolder,
		openDetail,
		openGit,
		loadPlugins,
		runPlugin,
		bindEvents,
	}
}
