import { ref, computed, watch, onMounted } from 'vue'
import { listPackages, getPackageVersions, installPackage } from '@/api/store'
import { pickDir } from '@/api/settings'
import type { PackageInfo, PackageVersion } from '@/types'

// 分类选项表
export const categories = [
	{ key: 'all', label: '全部' },
	{ key: 'language', label: '语言运行时' },
	{ key: 'webserver', label: 'Web 服务器' },
	{ key: 'database', label: '数据库' },
	{ key: 'cache', label: '缓存' }
]

// 镜像源选项表
export const mirrors = [
	{ key: 'aliyun', label: '阿里云镜像 (国内加速)' },
	{ key: 'official', label: '官方源' },
	{ key: 'tencent', label: '腾讯云镜像' }
]

// 搜索关键词
const keyword = ref('')

// 当前选中的分类键
const curCat = ref('all')

// 软件包全量列表
const packages = ref<PackageInfo[]>([])

// 远程同步状态
const syncing = ref(false)

// 安装弹窗可见性
const showModal = ref(false)

// 当前选中的软件包
const curPkg = ref<PackageInfo | null>(null)

// 当前软件包可用版本列表
const versions = ref<PackageVersion[]>([])

// 当前选中的版本
const curVer = ref<PackageVersion | null>(null)

// 当前选中的下载镜像源
const curMirror = ref('aliyun')

// 自定义安装目录路径
const customDir = ref('')

// 版本列表加载状态
const loadingVer = ref(false)

// 正在安装状态
const installing = ref(false)

// 安装进度百分比
const progress = ref(0)

// 过滤后的软件包列表
const filtered = computed(() => {
	let list = packages.value
	if (curCat.value !== 'all') {
		list = list.filter((p) => p.category === curCat.value)
	}
	if (keyword.value.trim()) {
		const q = keyword.value.toLowerCase().trim()
		list = list.filter(
			(p) =>
				p.displayName.toLowerCase().includes(q) ||
				p.description.toLowerCase().includes(q)
		)
	}
	return list
})

// 默认推导的安装目录路径
const defaultDir = computed(() => {
	if (!curPkg.value || !curVer.value) {
		return 'runtimes/'
	}
	const p = curPkg.value.key
	const v = curVer.value.version
	const arch = curVer.value.architecture
	if (curVer.value.variant) {
		return `runtimes\\${p}\\${v}-${curVer.value.variant}-${arch}`
	}
	return `runtimes\\${p}\\${v}-${arch}`
})

// 监听默认路径变化并同步到输入框
watch(defaultDir, (val) => {
	customDir.value = val
})

// 判断指定版本是否已安装
function isVerInstalled(ver: string): boolean {
	if (!curPkg.value || !ver) {
		return false
	}
	return curPkg.value.installedVersions.includes(ver)
}

// 选择安装目标目录
async function pickInstallDir() {
	const dir = await pickDir('选择软件安装目标目录')
	if (dir) {
		customDir.value = dir
	}
}

// 加载所有软件包
async function loadPackages() {
	try {
		packages.value = await listPackages()
	} catch (err) {
		toast.error(String(err))
	}
}

// 同步远程版本列表
async function syncPackages() {
	syncing.value = true
	try {
		await loadPackages()
		toast.success('已同步所有环境的最新远程版本列表')
	} catch (err) {
		toast.error(String(err))
	} finally {
		syncing.value = false
	}
}

// 应用推荐版本或首个版本
function pickDefaultVer(list: PackageVersion[]) {
	const rec = list.find((v) => v.isRecommended)
	if (rec) {
		curVer.value = rec
	} else if (list.length > 0) {
		curVer.value = list[0]
	} else {
		curVer.value = null
	}
}

// 打开软件包版本选择弹窗
async function openDetail(pkg: PackageInfo) {
	curPkg.value = pkg
	installing.value = false
	progress.value = 0
	showModal.value = true
	loadingVer.value = true

	try {
		const list = await getPackageVersions(pkg.key)
		versions.value = list
		pickDefaultVer(list)
	} catch (err) {
		toast.warn(String(err))
	} finally {
		loadingVer.value = false
	}
}

// 在线刷新当前软件包版本列表
async function refreshVer() {
	if (!curPkg.value) {
		return
	}
	loadingVer.value = true
	try {
		const list = await getPackageVersions(curPkg.value.key)
		versions.value = list
		pickDefaultVer(list)
		toast.success(`已在线更新 ${curPkg.value.displayName} 最新版本列表`)
	} catch (err) {
		toast.error(String(err))
	} finally {
		loadingVer.value = false
	}
}

// 开始安装选中的软件包版本
async function startInstall() {
	if (!curPkg.value || !curVer.value) {
		return
	}

	installing.value = true
	progress.value = 10

	let offProgress: (() => void) | null = null
	try {
		const wailsRt = await import('../../../../wailsjs/runtime/runtime')
		offProgress = wailsRt.EventsOn('task:progress', (ev: any) => {
			if (ev && typeof ev.progress === 'number') {
				progress.value = ev.progress
			}
		})
	} catch {
		// Ignore if not in Wails runtime environment
	}

	try {
		let targetDir = customDir.value.trim()
		if (!targetDir) {
			targetDir = defaultDir.value
		}
		const pkgName = curPkg.value.displayName
		const verStr = curVer.value.version
		const res = await installPackage({
			packageKey: curPkg.value.key,
			version: verStr,
			variant: curVer.value.variant,
			architecture: curVer.value.architecture,
			downloadUrl: curVer.value.downloadUrl,
			installPath: targetDir,
			mirror: curMirror.value
		})

		if (offProgress) {
			offProgress()
		}
		progress.value = 100

		setTimeout(async () => {
			if (res.serviceCreated) {
				toast.success(`${pkgName} ${verStr} 安装成功，已自动创建服务实例`)
			} else if (res.serviceId) {
				toast.success(`${pkgName} ${verStr} 安装成功，可在服务实例中切换到此版本`)
			} else {
				toast.success(`${pkgName} ${verStr} 安装成功`)
			}
			installing.value = false
			showModal.value = false
			await loadPackages()
		}, 300)
	} catch (err) {
		if (offProgress) {
			offProgress()
		}
		installing.value = false
		toast.error(String(err))
	}
}

// 导出软件商店逻辑
export function useStore() {
	onMounted(loadPackages)
	return {
		keyword,
		curCat,
		filtered,
		syncing,
		showModal,
		curPkg,
		versions,
		curVer,
		curMirror,
		customDir,
		loadingVer,
		installing,
		progress,
		isVerInstalled,
		pickInstallDir,
		syncPackages,
		openDetail,
		refreshVer,
		startInstall
	}
}

