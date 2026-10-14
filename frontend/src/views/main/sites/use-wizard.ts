import { ref, computed } from 'vue'
import { createSite } from '@/api/sites'
import { pickDir, pickFile } from '@/api/settings'
import type { SiteKind } from '@/types'
import { runtimes, services, loadAll } from './use-sites'

// 向导步骤名称列表
export const stepLabels = ['选择类型', '基本信息', '运行环境', '高级配置']

// 向导网站类型卡片配置
export const wizardKinds: Array<{
	kind: SiteKind
	title: string
	mode: string
	examples: string
	icon: string
}> = [
	{
		kind: 'php',
		title: 'PHP 网站',
		mode: 'Nginx + PHP-FPM',
		examples: 'Laravel / WordPress / ThinkPHP',
		icon: 'pi pi-code'
	},
	{
		kind: 'nodejs',
		title: 'Node.js 应用',
		mode: 'Nginx 反向代理',
		examples: 'Express / Nest.js / Next / Nuxt',
		icon: 'pi pi-bolt'
	},
	{
		kind: 'static',
		title: '静态网站',
		mode: 'Nginx 直接代理',
		examples: 'HTML / Vue / React / VitePress',
		icon: 'pi pi-file'
	},
	{
		kind: 'go',
		title: 'Go 应用',
		mode: 'Nginx 反向代理',
		examples: 'Gin / Fiber / Echo 二进制服务',
		icon: 'pi pi-box'
	},
	{
		kind: 'python',
		title: 'Python 应用',
		mode: 'Nginx + WSGI/ASGI',
		examples: 'FastAPI / Django / Flask',
		icon: 'pi pi-compass'
	}
]

// 向导弹窗可见性
const showWizard = ref(false)

// 当前向导步骤序号
const step = ref(1)

// 网站目录创建来源模式
const dirMode = ref<'new' | 'existing' | 'git'>('new')

// 表单提交加载状态
const submitting = ref(false)

// 创建网站表单对象
const form = ref({
	name: '',
	kind: 'php' as SiteKind,
	domains: ['localhost'],
	root: 'G:\\www\\',
	runtimeId: undefined as number | undefined,
	serviceId: undefined as number | undefined,
	autoService: true,
	upstream: 9000,
	port: 80,
	remote: '',
	branch: 'main',
	entry: 'index.php',
	cmd: '',
	tls: false,
	issuer: 'self-signed',
	cert: '',
	key: '',
	autoStart: true,
	startNow: true,
	custom: ''
})

// 当前向导步骤标题
const stepTitle = computed(() => stepLabels[step.value - 1])

// 匹配当前网站类型的运行环境列表
const matchingRuntimes = computed(() => {
	if (form.value.kind === 'php') {
		return runtimes.value.filter((r) => r.runtimeType === 'php')
	}
	return runtimes.value
})

// 服务实例下拉选项
const serviceOptions = computed(() =>
	services.value.map((s) => ({
		label: `${s.name} (#${s.id})`,
		value: s.id
	}))
)

// 启动命令占位提示文本
const cmdHint = computed(() => {
	if (form.value.kind === 'nodejs') {
		return 'npm start 或 node app.js'
	}
	if (form.value.kind === 'go') {
		return './main.exe'
	}
	if (form.value.kind === 'python') {
		return 'uvicorn main:app --port 8000'
	}
	return ''
})

// 打开添加网站向导弹窗
function openWizard() {
	step.value = 1
	dirMode.value = 'new'
	const defaultPhp = runtimes.value.find((r) => r.runtimeType === 'php')
	let defaultRtId: number | undefined
	if (defaultPhp) {
		defaultRtId = defaultPhp.id
	}
	form.value = {
		name: '',
		kind: 'php',
		domains: ['localhost'],
		root: 'G:\\www\\',
		runtimeId: defaultRtId,
		serviceId: undefined,
		autoService: true,
		upstream: 9000,
		port: 80,
		remote: '',
		branch: 'main',
		entry: 'index.php',
		cmd: '',
		tls: false,
		issuer: 'self-signed',
		cert: '',
		key: '',
		autoStart: true,
		startNow: true,
		custom: ''
	}
	showWizard.value = true
}

// 跳转到已完成的向导步骤
function jumpStep(s: number) {
	if (s < step.value) {
		step.value = s
	}
}

// 选择网站类型并填充默认端口与命令
function selectKind(kind: SiteKind) {
	form.value.kind = kind
	if (kind === 'static') {
		form.value.entry = 'index.html'
	} else if (kind === 'php') {
		form.value.entry = 'index.php'
		form.value.upstream = 9000
	} else if (kind === 'nodejs') {
		form.value.upstream = 3000
		form.value.cmd = 'npm start'
	} else if (kind === 'go') {
		form.value.upstream = 8080
		form.value.cmd = './main.exe'
	} else if (kind === 'python') {
		form.value.upstream = 8000
		form.value.cmd = 'app:application'
	}
}

// 输入网站名称时自动推导本地目录（默认域名保持 localhost）
function onNameInput() {
	if (!form.value.domains[0]) {
		form.value.domains[0] = 'localhost'
	}
	const slug = form.value.name
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9-_]/g, '')
	if (slug && form.value.root.startsWith('G:\\www\\')) {
		form.value.root = `G:\\www\\${slug}`
	}
}

// 校验并进入下一步
function nextStep() {
	if (step.value === 2) {
		if (!form.value.name.trim()) {
			toast.warn('请输入网站名称')
			return
		}
		const validDomains = form.value.domains.map((d) => d.trim()).filter(Boolean)
		if (validDomains.length === 0) {
			toast.warn('请至少填写一个绑定的本地域名')
			return
		}
		if (!form.value.root.trim()) {
			toast.warn('请填写网站根目录路径')
			return
		}
	}
	step.value++
}

// 选择创建网站的物理根目录
async function pickCreateRoot() {
	const dir = await pickDir('选择网站根目录')
	if (!dir) {
		return
	}
	form.value.root = dir
	if (!form.value.name.trim()) {
		const parts = dir.replace(/\\/g, '/').split('/').filter(Boolean)
		if (parts.length > 0) {
			form.value.name = parts[parts.length - 1]
			onNameInput()
			form.value.root = dir
		}
	}
}

// 选择创建网站的 SSL 证书文件
async function pickCreateCert() {
	const file = await pickFile('选择 SSL 证书文件 (.pem / .crt)')
	if (file) {
		form.value.cert = file
	}
}

// 选择创建网站的 SSL 私钥文件
async function pickCreateKey() {
	const file = await pickFile('选择 SSL 私钥文件 (.key)')
	if (file) {
		form.value.key = file
	}
}

// 提交创建新网站
async function submitCreate() {
	const cleanDomains = form.value.domains.map((d) => d.trim()).filter(Boolean)
	if (!form.value.name.trim() || cleanDomains.length === 0) {
		toast.warn('请完善网站名称与域名绑定')
		return
	}

	submitting.value = true
	try {
		let gitEnabled = false
		if (dirMode.value === 'git' || Boolean(form.value.remote.trim())) {
			gitEnabled = true
		}
		await createSite({
			name: form.value.name.trim(),
			kind: form.value.kind,
			domains: cleanDomains,
			root: form.value.root.trim(),
			runtimeId: form.value.runtimeId,
			serviceId: form.value.serviceId,
			autoService: form.value.autoService,
			upstream: form.value.upstream,
			git: gitEnabled,
			remote: form.value.remote.trim(),
			branch: form.value.branch.trim(),
			port: form.value.port,
			tls: form.value.tls,
			issuer: form.value.issuer,
			cert: form.value.cert,
			key: form.value.key,
			entry: form.value.entry,
			cmd: form.value.cmd,
			autoStart: form.value.autoStart,
			startNow: form.value.startNow,
			custom: form.value.custom
		})
		toast.success(`网站「${form.value.name}」已成功创建`)
		showWizard.value = false
		await loadAll()
	} catch (err) {
		toast.error(String(err))
	} finally {
		submitting.value = false
	}
}

// 导出添加网站向导逻辑
export function useWizard() {
	return {
		showWizard,
		step,
		stepTitle,
		dirMode,
		submitting,
		form,
		matchingRuntimes,
		serviceOptions,
		cmdHint,
		openWizard,
		jumpStep,
		selectKind,
		onNameInput,
		nextStep,
		pickCreateRoot,
		pickCreateCert,
		pickCreateKey,
		submitCreate
	}
}
