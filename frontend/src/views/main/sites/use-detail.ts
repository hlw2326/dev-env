import { ref, computed } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import {
	updateSite,
	getSiteConf,
	renderSiteConf,
	saveSiteConf,
	checkSiteConf,
	getSiteLogs,
	getGitStatus,
	pullGit,
	commitGit,
	pushGit
} from '@/api/sites'
import { pickDir, pickFile } from '@/api/settings'
import type { Site, GitStatus, SiteSnapshot } from '@/types'
import { sites, runtimes, loadAll } from './use-sites'

// 详情页签类型定义（仅保留有真实后端支撑的 5 个页签）
export type DetailTab = 'info' | 'git' | 'conf' | 'ssl' | 'logs'

// Nginx 配置模板下拉列表
export const nginxTpls = [
	{ key: 'cache', label: '静态文件缓存 (30天)' },
	{ key: 'gzip', label: 'Gzip 压缩配置' },
	{ key: 'cors', label: 'CORS 跨域配置' },
	{ key: 'websocket', label: 'WebSocket 反向代理' },
	{ key: 'https_redirect', label: 'HTTP 强制跳转 HTTPS' },
	{ key: 'laravel_rewrite', label: 'Laravel / ThinkPHP 伪静态' },
	{ key: 'security', label: '安全防护响应头' }
]

// Nginx 模板代码片段映射
const tplSnippets: Record<string, string> = {
	cache: `\n    # 静态文件缓存\n    location ~* \\.(jpg|jpeg|png|gif|ico|css|js|woff2|svg)$ {\n        expires 30d;\n        add_header Cache-Control "public, immutable";\n        access_log off;\n    }\n`,
	gzip: `\n    # Gzip 压缩\n    gzip on;\n    gzip_vary on;\n    gzip_min_length 1024;\n    gzip_types text/plain text/css application/json application/javascript text/xml;\n`,
	cors: `\n    # CORS 跨域配置\n    add_header Access-Control-Allow-Origin *;\n    add_header Access-Control-Allow-Methods 'GET, POST, OPTIONS, PUT, DELETE';\n    add_header Access-Control-Allow-Headers 'DNT,Keep-Alive,User-Agent,X-Requested-With,Content-Type,Authorization';\n`,
	websocket: `\n    # WebSocket 反向代理\n    location /ws {\n        proxy_pass http://127.0.0.1:3000;\n        proxy_http_version 1.1;\n        proxy_set_header Upgrade $http_upgrade;\n        proxy_set_header Connection "upgrade";\n        proxy_set_header Host $host;\n    }\n`,
	https_redirect: `\n    # HTTP 强制跳转 HTTPS\n    if ($scheme != "https") {\n        return 301 https://$host$request_uri;\n    }\n`,
	laravel_rewrite: `\n    # Laravel / ThinkPHP 伪静态\n    location / {\n        try_files $uri $uri/ /index.php?$query_string;\n    }\n`,
	security: `\n    # 安全头部配置\n    server_tokens off;\n    add_header X-Frame-Options "SAMEORIGIN" always;\n    add_header X-XSS-Protection "1; mode=block" always;\n    add_header X-Content-Type-Options "nosniff" always;\n`
}

// 详情弹窗可见性
const showDetail = ref(false)

// 当前查看的网站实例
const curSite = ref<Site | null>(null)

// 当前激活的详情页签
const curTab = ref<DetailTab>('info')

// 保存提交状态
const saving = ref(false)

// 网站基本与 SSL 编辑表单
const editForm = ref({
	name: '',
	port: 80,
	root: '',
	entry: 'index.php',
	upstream: 9000,
	runtimeId: undefined as number | undefined,
	autoStart: true,
	tls: false,
	issuer: 'self-signed',
	cert: '',
	key: ''
})

// 域名多行编辑文本
const domainsText = ref('')

// 运行环境选择下拉列表
const runtimeOpts = computed(() => {
	const opts: Array<{ label: string; value: number | undefined }> = [
		{ label: '不指定 / 默认环境', value: undefined }
	]
	for (const r of runtimes.value) {
		opts.push({
			label: `${r.runtimeType.toUpperCase()} ${r.version} (${r.architecture})`,
			value: r.id
		})
	}
	return opts
})

// Git 仓库状态
const gitStatus = ref<GitStatus | null>(null)

// Git 操作加载状态
const gitLoading = ref(false)

// 勾选待提交的文件列表
const selectedFiles = ref<string[]>([])

// Git 提交信息
const commitMsg = ref('')

// Nginx 配置子模式
const nginxMode = ref<'editor' | 'visual' | 'history'>('editor')

// Nginx 配置文件文本内容
const confContent = ref('')

// Nginx 配置语法检测结果
const confCheck = ref<{ valid: boolean; message: string } | null>(null)

// 选中的快捷插入模板
const selectedTpl = ref<string | null>(null)

// 可视化 Nginx 规则表单
const visualForm = ref({
	gzip: true,
	staticCache: true,
	cors: false,
	securityHeaders: true,
	ipPolicy: 'allow_all',
	ipList: ''
})

// 日志子模式
const logKind = ref<'access' | 'error' | 'stats'>('access')

// 网站日志行列表
const siteLogs = ref<string[]>([])

// 打开网站详情模态框
export async function openDetail(site: Site, tab: DetailTab = 'info') {
	curSite.value = site
	curTab.value = tab
	confCheck.value = null
	let upstreamVal = 9000
	if (site.upstream) {
		upstreamVal = site.upstream
	}
	editForm.value = {
		name: site.name,
		port: site.port,
		root: site.root,
		entry: site.meta.entry,
		upstream: upstreamVal,
		runtimeId: site.runtimeId,
		autoStart: site.meta.autoStart,
		tls: site.tls,
		issuer: site.issuer,
		cert: site.cert,
		key: site.key
	}
	domainsText.value = site.domains.join('\n')
	showDetail.value = true
	await switchTab(tab)
}

// 切换详情页签并按需加载数据
export async function switchTab(tab: DetailTab) {
	curTab.value = tab
	if (!curSite.value) {
		return
	}
	if (tab === 'git') {
		await loadGit(curSite.value.id)
	} else if (tab === 'conf') {
		confContent.value = await getSiteConf(curSite.value.id)
	} else if (tab === 'logs') {
		await loadLogs(curSite.value.id)
	}
}

// 保存网站基本信息或 SSL 配置
async function saveInfo() {
	if (!curSite.value) {
		return
	}
	const domains = domainsText.value
		.split(/[\n,]/)
		.map((d) => d.trim())
		.filter(Boolean)
	if (domains.length === 0) {
		toast.warn('请至少保留一个绑定域名')
		return
	}
	saving.value = true
	try {
		const updated = await updateSite({
			id: curSite.value.id,
			name: editForm.value.name,
			domains,
			root: editForm.value.root,
			runtimeId: editForm.value.runtimeId,
			port: editForm.value.port,
			tls: editForm.value.tls,
			issuer: editForm.value.issuer,
			cert: editForm.value.cert,
			key: editForm.value.key,
			upstream: editForm.value.upstream,
			entry: editForm.value.entry,
			autoStart: editForm.value.autoStart
		})
		curSite.value = updated
		toast.success('网站配置已保存并同步更新 Nginx 虚拟主机')
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		saving.value = false
	}
}

// 选择网站根目录
async function pickEditRoot() {
	const dir = await pickDir('选择网站根目录')
	if (dir) {
		editForm.value.root = dir
	}
}

// 选择 SSL 证书文件
async function pickEditCert() {
	const file = await pickFile('选择 SSL 证书文件 (.pem / .crt)')
	if (file) {
		editForm.value.cert = file
	}
}

// 选择 SSL 私钥文件
async function pickEditKey() {
	const file = await pickFile('选择 SSL 私钥文件 (.key)')
	if (file) {
		editForm.value.key = file
	}
}

// 加载网站 Git 仓库状态
async function loadGit(siteId: number) {
	gitLoading.value = true
	try {
		gitStatus.value = await getGitStatus(siteId)
		selectedFiles.value = [
			...gitStatus.value.modified,
			...gitStatus.value.untracked
		]
	} catch (err) {
		toast.error(String(err))
	} finally {
		gitLoading.value = false
	}
}

// 拉取远程 Git 仓库
async function pullRepo() {
	if (!curSite.value) {
		return
	}
	gitLoading.value = true
	try {
		await pullGit(curSite.value.id)
		toast.success('已从远程仓库拉取最新代码')
		await loadGit(curSite.value.id)
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		gitLoading.value = false
	}
}

// 提交本地更改并可选推送到远程
async function commitRepo(andPush: boolean) {
	if (!curSite.value || !commitMsg.value.trim()) {
		return
	}
	gitLoading.value = true
	try {
		await commitGit(curSite.value.id, commitMsg.value.trim(), selectedFiles.value)
		if (andPush) {
			await pushGit(curSite.value.id)
			toast.success('代码已成功提交并推送到远程仓库')
		} else {
			toast.success('代码更改已提交到本地仓库')
		}
		commitMsg.value = ''
		await loadGit(curSite.value.id)
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		gitLoading.value = false
	}
}

// 推送本地提交到远程分支
async function pushRepo() {
	if (!curSite.value) {
		return
	}
	gitLoading.value = true
	try {
		await pushGit(curSite.value.id)
		toast.success('已成功推送到远程分支')
		await loadGit(curSite.value.id)
	} catch (err) {
		toast.error(String(err))
	} finally {
		gitLoading.value = false
	}
}

// 插入 Nginx 预设配置片段
function insertTpl() {
	const key = selectedTpl.value
	if (!key) {
		return
	}
	const snippet = tplSnippets[key]
	if (snippet) {
		const lastBrace = confContent.value.lastIndexOf('}')
		if (lastBrace !== -1) {
			confContent.value = confContent.value.slice(0, lastBrace) + snippet + '}\n'
		} else {
			confContent.value += snippet
		}
		toast.info('已插入配置模板片段')
	}
	selectedTpl.value = null
}

// 应用可视化规则到 Nginx 配置
async function applyVisual() {
	if (!curSite.value) {
		return
	}
	let extra = ''
	if (visualForm.value.gzip) {
		extra += `gzip on;\ngzip_comp_level 6;\ngzip_types text/plain text/css application/json application/javascript;\n`
	}
	if (visualForm.value.staticCache) {
		extra += `location ~* \\.(jpg|jpeg|png|gif|ico|css|js|svg|woff2)$ {\n    expires 30d;\n    access_log off;\n}\n`
	}
	if (visualForm.value.cors) {
		extra += `add_header Access-Control-Allow-Origin *;\n`
	}
	if (visualForm.value.securityHeaders) {
		extra += `server_tokens off;\nadd_header X-Frame-Options "SAMEORIGIN" always;\nadd_header X-Content-Type-Options "nosniff" always;\n`
	}
	if (visualForm.value.ipPolicy !== 'allow_all' && visualForm.value.ipList.trim()) {
		const ips = visualForm.value.ipList
			.split('\n')
			.map((i) => i.trim())
			.filter(Boolean)
		if (visualForm.value.ipPolicy === 'whitelist') {
			ips.forEach((ip) => (extra += `allow ${ip};\n`))
			extra += `deny all;\n`
		} else {
			ips.forEach((ip) => (extra += `deny ${ip};\n`))
			extra += `allow all;\n`
		}
	}
	const base = await renderSiteConf(curSite.value.id)
	const lastBrace = base.lastIndexOf('}')
	let block = ''
	if (extra) {
		block = `\n    # 可视化配置生成\n    ${extra.trim().replace(/\n/g, '\n    ')}\n`
	}
	confContent.value = base.slice(0, lastBrace) + block + '}\n'
	nginxMode.value = 'editor'
	toast.success('已根据表单规则更新配置文件内容，请点击保存生效')
}

// 测试当前 Nginx 配置语法
async function testConf() {
	confCheck.value = await checkSiteConf(confContent.value)
}

// 保存并重载 Nginx 配置
async function saveConf() {
	if (!curSite.value) {
		return
	}
	saving.value = true
	try {
		await saveSiteConf(curSite.value.id, confContent.value, '在线编辑 Nginx 配置')
		toast.success('Nginx 配置已保存并成功重载')
		await loadAll()
		const refreshed = sites.value.find((s) => s.id === curSite.value?.id)
		if (refreshed) {
			curSite.value = refreshed
		}
	} catch (err) {
		toast.error(String(err))
	} finally {
		saving.value = false
	}
}

// 生成本地自签名证书路径
function genSelfCert() {
	if (!curSite.value) {
		return
	}
	const primary = curSite.value.domains[0]
	editForm.value.tls = true
	editForm.value.issuer = 'self-signed'
	editForm.value.cert = `D:\\DevEnv\\certs\\${primary}-selfsigned.pem`
	editForm.value.key = `D:\\DevEnv\\certs\\${primary}-selfsigned.key`
	toast.success(`已为 ${primary} 生成本地自签名 SSL 证书`)
}

// 配置 Let's Encrypt 证书路径
function reqLetsCert() {
	if (!curSite.value) {
		return
	}
	const primary = curSite.value.domains[0]
	editForm.value.tls = true
	editForm.value.issuer = 'letsencrypt'
	editForm.value.cert = `D:\\DevEnv\\certs\\letsencrypt\\${primary}\\fullchain.pem`
	editForm.value.key = `D:\\DevEnv\\certs\\letsencrypt\\${primary}\\privkey.pem`
	toast.success(`已为 ${primary} 配置 Let's Encrypt 证书路径`)
}

// 切换日志类型并加载
async function switchLog(kind: 'access' | 'error') {
	logKind.value = kind
	if (curSite.value) {
		await loadLogs(curSite.value.id)
	}
}

// 加载网站访问或错误日志
async function loadLogs(siteId: number) {
	if (logKind.value === 'stats') {
		return
	}
	siteLogs.value = await getSiteLogs(siteId, logKind.value, 100)
}

// 导出网站详情弹窗状态与方法
export function useDetail() {
	// 气泡确认框实例
	const confirm = useConfirm()

	// 重置为默认生成的 Nginx 配置
	function resetConf(event: Event) {
		confirm.require({
			target: event.currentTarget as HTMLElement,
			message: '确定要将当前编辑内容恢复为默认生成的 Nginx 配置吗？',
			icon: 'pi pi-exclamation-triangle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: '恢复默认',
				severity: 'danger',
				size: 'small'
			},
			accept: async () => {
				if (!curSite.value) {
					return
				}
				confContent.value = await renderSiteConf(curSite.value.id)
				confCheck.value = null
				toast.info('已恢复默认生成配置，请点击保存生效')
			}
		})
	}

	// 恢复指定历史快照配置
	function restoreSnap(event: Event, snap: SiteSnapshot) {
		confirm.require({
			target: event.currentTarget as HTMLElement,
			message: `确定要恢复 ${snap.timestamp} 的配置版本吗？`,
			icon: 'pi pi-info-circle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: '恢复此版本',
				size: 'small'
			},
			accept: () => {
				confContent.value = snap.content
				nginxMode.value = 'editor'
				toast.info(`已加载 ${snap.timestamp} 的配置版本，点击保存可立即生效`)
			}
		})
	}

	return {
		showDetail,
		curSite,
		curTab,
		saving,
		editForm,
		domainsText,
		runtimeOpts,
		gitStatus,
		gitLoading,
		selectedFiles,
		commitMsg,
		nginxMode,
		confContent,
		confCheck,
		selectedTpl,
		visualForm,
		logKind,
		siteLogs,
		openDetail,
		switchTab,
		saveInfo,
		pickEditRoot,
		pickEditCert,
		pickEditKey,
		loadGit,
		pullRepo,
		commitRepo,
		pushRepo,
		insertTpl,
		applyVisual,
		resetConf,
		testConf,
		saveConf,
		restoreSnap,
		genSelfCert,
		reqLetsCert,
		switchLog,
		loadLogs
	}
}
