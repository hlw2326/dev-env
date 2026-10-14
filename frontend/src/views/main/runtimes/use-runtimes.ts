import { ref, computed, onMounted } from 'vue'
import {
	listAllRuntimes,
	registerRuntime,
	uninstallRuntime,
	setDefaultVersion,
	healthCheck,
	getExtensions
} from '@/api/runtimes'
import { pickDir } from '@/api/settings'
import type { RuntimeVersion, PHPExtension } from '@/types'

// 环境类型名称映射表
const typeNames: Record<string, string> = {
	php: 'PHP',
	nginx: 'Nginx',
	mysql: 'MySQL',
	postgresql: 'PostgreSQL',
	redis: 'Redis',
	apache: 'Apache',
	nodejs: 'Node.js'
}

// 环境类型图标映射表
const typeIcons: Record<string, string> = {
	php: 'pi pi-code',
	nginx: 'pi pi-server',
	mysql: 'pi pi-database',
	redis: 'pi pi-bolt',
	postgresql: 'pi pi-database',
	apache: 'pi pi-compass',
	nodejs: 'pi pi-code'
}

// 获取环境类型显示名
export function getTypeName(type: string): string {
	const found = typeNames[type.toLowerCase()]
	if (found) {
		return found
	}
	return type.toUpperCase()
}

// 获取环境类型图标类名
export function getTypeIcon(type: string): string {
	const found = typeIcons[type.toLowerCase()]
	if (found) {
		return found
	}
	return 'pi pi-box'
}

// 列表加载状态
const loading = ref(false)

// 当前正在操作版本的编号
const busyId = ref<number | null>(null)

// 已安装运行环境全量列表
const runtimes = ref<RuntimeVersion[]>([])

// 展开的分组列表
const openGroups = ref<string[]>(['php', 'nginx', 'mysql', 'redis', 'postgresql', 'nodejs'])

// 扩展管理弹窗可见性
const showExt = ref(false)

// 当前选中的运行时实例
const curRt = ref<RuntimeVersion | null>(null)

// PHP 扩展列表
const extensions = ref<PHPExtension[]>([])

// 扩展加载状态
const loadingExt = ref(false)

// 卸载确认弹窗可见性
const showUninstall = ref(false)

// 是否同时删除物理目录
const purgeFiles = ref(false)

// 正在卸载状态
const uninstalling = ref(false)

// 导入本地环境弹窗可见性
const showImport = ref(false)

// 正在导入状态
const importing = ref(false)

// 导入本地环境表单
const importForm = ref({
	kind: 'php',
	version: '',
	arch: 'x64',
	dir: ''
})

// 按环境类型分组的字典
const grouped = computed(() => {
	const map: Record<string, RuntimeVersion[]> = {}
	runtimes.value.forEach((rt) => {
		const t = rt.runtimeType.toLowerCase()
		if (!map[t]) {
			map[t] = []
		}
		map[t].push(rt)
	})
	return map
})

// 判断分组是否展开
function isGroupOpen(type: string): boolean {
	return openGroups.value.includes(type)
}

// 切换分组展开或收起
function toggleGroup(type: string) {
	const idx = openGroups.value.indexOf(type)
	if (idx > -1) {
		openGroups.value.splice(idx, 1)
	} else {
		openGroups.value.push(type)
	}
}

// 加载所有运行环境列表
async function loadRuntimes() {
	loading.value = true
	try {
		runtimes.value = await listAllRuntimes()
	} catch (err) {
		toast.error(String(err))
	} finally {
		loading.value = false
	}
}

// 设为默认全局 CLI 版本
async function setDefault(rt: RuntimeVersion) {
	busyId.value = rt.id
	try {
		await setDefaultVersion(rt.id)
		toast.success(`已将 PHP ${rt.version} 设为默认全局 CLI`)
		await loadRuntimes()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 执行运行时健康检查
async function checkHealth(rt: RuntimeVersion) {
	try {
		const res = await healthCheck(rt.id)
		if (res.healthy) {
			toast.success(res.message)
		} else {
			toast.warn(res.message)
		}
	} catch (err) {
		toast.error(String(err))
	}
}

// 打开 PHP 扩展管理弹窗
async function openExt(rt: RuntimeVersion) {
	curRt.value = rt
	showExt.value = true
	loadingExt.value = true
	try {
		extensions.value = await getExtensions(rt.id)
	} catch (err) {
		toast.error(String(err))
	} finally {
		loadingExt.value = false
	}
}

// 打开卸载确认弹窗
function openUninstall(rt: RuntimeVersion) {
	curRt.value = rt
	purgeFiles.value = false
	showUninstall.value = true
}

// 确认卸载运行时版本
async function confirmUninstall() {
	if (!curRt.value) {
		return
	}
	uninstalling.value = true
	try {
		const res = await uninstallRuntime(curRt.value.id, purgeFiles.value)
		let msg = '运行时已卸载'
		if (res.message) {
			msg = res.message
		}
		toast.success(msg)
		showUninstall.value = false
		await loadRuntimes()
	} catch (err) {
		toast.error(String(err))
	} finally {
		uninstalling.value = false
	}
}

// 打开导入本地环境弹窗
function openImport() {
	importForm.value = {
		kind: 'php',
		version: '',
		arch: 'x64',
		dir: ''
	}
	showImport.value = true
}

// 选择本地环境安装目录
async function pickImportDir() {
	const dir = await pickDir('选择本地运行环境安装目录')
	if (!dir) {
		return
	}
	importForm.value.dir = dir
	if (!importForm.value.version.trim()) {
		const m = dir.match(/(\d+\.\d+(?:\.\d+)?)/)
		if (m) {
			importForm.value.version = m[1]
		}
	}
}

// 提交导入本地环境
async function submitImport() {
	if (!importForm.value.version.trim() || !importForm.value.dir.trim()) {
		toast.warn('请填写环境版本号并选择本地目录')
		return
	}
	importing.value = true
	try {
		await registerRuntime(
			importForm.value.kind,
			importForm.value.dir.trim(),
			importForm.value.version.trim(),
			importForm.value.arch
		)
		toast.success('本地运行环境已成功导入')
		showImport.value = false
		await loadRuntimes()
	} catch (err) {
		toast.error(String(err))
	} finally {
		importing.value = false
	}
}

// 当前选中运行时的同类型版本总数
const curRtSameTypeCount = computed(() => {
	if (!curRt.value) {
		return 1
	}
	const list = grouped.value[curRt.value.runtimeType.toLowerCase()]
	return list ? list.length : 1
})

// 导出运行环境管理逻辑
export function useRuntimes() {
	onMounted(loadRuntimes)

	return {
		loading,
		busyId,
		runtimes,
		grouped,
		showExt,
		curRt,
		curRtSameTypeCount,
		extensions,
		loadingExt,
		showUninstall,
		purgeFiles,
		uninstalling,
		showImport,
		importing,
		importForm,
		isGroupOpen,
		toggleGroup,
		loadRuntimes,
		setDefault,
		checkHealth,
		openExt,
		openUninstall,
		confirmUninstall,
		openImport,
		pickImportDir,
		submitImport
	}
}

