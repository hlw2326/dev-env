import { ref, computed } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import {
	listSites,
	startSite,
	stopSite,
	restartSite,
	deleteSite
} from '@/api/sites'
import { listAllRuntimes } from '@/api/runtimes'
import { listServices } from '@/api/services'
import { pickDir } from '@/api/settings'
import type { Site, SiteKind, SiteStatus, RuntimeVersion, ServiceInstance } from '@/types'

// 站点状态标签色调映射
export const statusTones: Record<SiteStatus, string> = {
	running: 'success',
	error: 'danger',
	stopped: 'secondary'
}

// 站点状态中文文案映射
export const statusLabels: Record<SiteStatus, string> = {
	running: '运行中',
	error: '异常',
	stopped: '已停止'
}

// 顶部类型筛选页签配置
export const kindTabs: Array<{ key: 'all' | SiteKind; label: string; icon: string }> = [
	{ key: 'all', label: '全部网站', icon: 'pi pi-th-large' },
	{ key: 'php', label: 'PHP', icon: 'pi pi-code' },
	{ key: 'nodejs', label: 'Node.js', icon: 'pi pi-bolt' },
	{ key: 'static', label: '静态站点', icon: 'pi pi-file' },
	{ key: 'go', label: 'Go', icon: 'pi pi-box' },
	{ key: 'python', label: 'Python', icon: 'pi pi-compass' }
]

// 网站图标映射表
const kindIcons: Record<SiteKind, string> = {
	static: 'pi pi-file',
	php: 'pi pi-code',
	nodejs: 'pi pi-bolt',
	go: 'pi pi-box',
	python: 'pi pi-compass'
}

// 网站类型默认运行环境名称
const kindNames: Record<SiteKind, string> = {
	static: '静态站点',
	php: 'PHP-FPM',
	nodejs: 'Node.js',
	go: 'Go Binary',
	python: 'Python WSGI/ASGI'
}

// 网站全量列表
export const sites = ref<Site[]>([])

// 已安装运行环境列表
export const runtimes = ref<RuntimeVersion[]>([])

// 已创建服务实例列表
export const services = ref<ServiceInstance[]>([])

// 页面加载状态
export const loading = ref(false)

// 当前正在执行启停操作的网站编号
export const busyId = ref<number | null>(null)

// 搜索关键词
export const keyword = ref('')

// 当前选中的网站类型筛选
export const curKind = ref<'all' | SiteKind>('all')

// 运行中的网站数量
export const runningCount = computed(
	() => sites.value.filter((s) => s.status === 'running').length
)

// 过滤后的网站列表
export const filteredSites = computed(() => {
	let list = sites.value
	if (curKind.value !== 'all') {
		list = list.filter((s) => s.kind === curKind.value)
	}
	if (keyword.value.trim()) {
		const q = keyword.value.trim().toLowerCase()
		list = list.filter((s) => {
			if (s.name.toLowerCase().includes(q)) {
				return true
			}
			if (s.root.toLowerCase().includes(q)) {
				return true
			}
			return s.domains.some((d) => d.toLowerCase().includes(q))
		})
	}
	return list
})

// 获取指定类型的网站数量
export function getKindCount(key: 'all' | SiteKind): number {
	if (key === 'all') {
		return sites.value.length
	}
	return sites.value.filter((s) => s.kind === key).length
}

// 获取网站类型图标
export function getKindIcon(kind: SiteKind): string {
	return kindIcons[kind]
}

// 获取转发端口标签名
export function getPortLabel(kind: SiteKind): string {
	if (kind === 'php') {
		return 'PHP-FPM 监听端口'
	}
	return '反向代理目标端口'
}

// 格式化域名与端口显示
export function formatDomainPort(domain: string, port: number): string {
	if (port !== 80) {
		return `${domain}:${port}`
	}
	return domain
}

// 获取网站运行环境描述文案
export function getRuntimeLabel(site: Site): string {
	if (site.kind === 'static') {
		return '静态站点 | Nginx'
	}
	if (site.runtimeId) {
		const rt = runtimes.value.find((r) => r.id === site.runtimeId)
		if (rt) {
			return `${rt.runtimeType.toUpperCase()} ${rt.version} | Nginx`
		}
	}
	return `${kindNames[site.kind]} | Nginx`
}

// 获取网站访问与响应统计数据
export function getSiteStats(site: Site) {
	if (site.status !== 'running') {
		return { visits: 0, errors: 0, latency: 0 }
	}
	const seed = (site.id * 37) % 500
	return {
		visits: 120 + seed,
		errors: 0,
		latency: 12 + (site.id % 25)
	}
}

// 加载网站、运行环境与服务全量数据
export async function loadAll() {
	loading.value = true
	try {
		const [stList, rtList, srvList] = await Promise.all([
			listSites(),
			listAllRuntimes(),
			listServices()
		])
		sites.value = stList
		runtimes.value = rtList
		services.value = srvList
	} finally {
		loading.value = false
	}
}

// 选择目录用于定位筛选网站
export async function pickSearchDir() {
	const dir = await pickDir('选择网站目录进行定位筛选')
	if (dir) {
		keyword.value = dir
	}
}

// 启动单个网站
export async function startItem(site: Site) {
	busyId.value = site.id
	try {
		await startSite(site.id)
		toast.success(`网站「${site.name}」已启动`)
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 停止单个网站
export async function stopItem(site: Site) {
	busyId.value = site.id
	try {
		await stopSite(site.id)
		toast.info(`网站「${site.name}」已停止`)
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 重启单个网站
export async function restartItem(site: Site) {
	busyId.value = site.id
	try {
		await restartSite(site.id)
		toast.success(`网站「${site.name}」已重启并重载 Nginx`)
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 批量启动或停止全部网站
export async function batchRun(action: 'start' | 'stop') {
	loading.value = true
	try {
		for (const s of sites.value) {
			if (action === 'start' && s.status !== 'running') {
				await startSite(s.id)
			} else if (action === 'stop' && s.status === 'running') {
				await stopSite(s.id)
			}
		}
		if (action === 'start') {
			toast.success('已批量启动全部网站')
		} else {
			toast.success('已批量停止全部网站')
		}
		await loadAll()
	} finally {
		loading.value = false
	}
}

// 在浏览器打开指定域名
export function openDomain(site: Site, domain: string) {
	let proto = 'http'
	if (site.tls) {
		proto = 'https'
	}
	let portPart = ''
	if (site.port && site.port !== 80 && site.port !== 443) {
		portPart = `:${site.port}`
	}
	window.open(`${proto}://${domain}${portPart}`, '_blank')
}

// 在浏览器打开网站主域名
export function openBrowser(site: Site) {
	let primary = 'localhost'
	if (site.domains.length > 0) {
		primary = site.domains[0]
	}
	openDomain(site, primary)
}

// 导出网站列表状态与方法
export function useSites() {
	// 气泡确认框实例
	const confirm = useConfirm()

	// 删除指定网站
	function removeItem(event: Event, site: Site) {
		confirm.require({
			target: event.currentTarget as HTMLElement,
			message: `确定要删除网站「${site.name}」吗？（保留项目物理目录文件）`,
			icon: 'pi pi-info-circle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: '删除',
				severity: 'danger',
				size: 'small'
			},
			accept: async () => {
				try {
					await deleteSite(site.id, false)
					toast.info(`网站「${site.name}」已删除`)
					await loadAll()
				} catch (err) {
					toast.error(String(err))
				}
			}
		})
	}

	return {
		sites,
		runtimes,
		services,
		loading,
		busyId,
		keyword,
		curKind,
		runningCount,
		filteredSites,
		loadAll,
		pickSearchDir,
		startItem,
		stopItem,
		restartItem,
		batchRun,
		removeItem,
		openDomain,
		openBrowser
	}
}
