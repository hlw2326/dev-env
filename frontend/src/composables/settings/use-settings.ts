import { ref, reactive, computed } from 'vue'
import {
	loadSettings,
	saveSettings,
	pickFolder,
	pickFile,
	openFolder,
	openNginxConfig as openNginxConfigApi,
	loadFolder,
} from '@/api'
import { useTheme, type ThemeMode } from '@/composables/theme'
import { toast, formatError } from '@/utils'

// 设置表单状态
const form = reactive({
	theme: 'light',
	root: '',
	port: 80,
	hosts: false,
	developerToolsCLI: '',
})

// 网关监听状态
const listening = ref(false)

// 网关错误消息
const nginxError = ref('')

// 应用数据目录路径
const dataFolder = ref('')

// 保存中状态
const saving = ref(false)

// 全局设置组合式函数
export function useSettings() {
	// 主题状态与操作函数
	const { theme, setMode, setColor, THEME_COLORS } = useTheme()

	// 当前强调色显示名称
	const colorName = computed(() => {
		const matched = THEME_COLORS.find((item) => item.value.toLowerCase() === theme.color.toLowerCase())
		if (matched) {
			return matched.name
		}
		return theme.color
	})

	// 加载后端全局设置与数据目录
	async function fetchData(): Promise<void> {
		try {
			const [settings, folder] = await Promise.all([loadSettings(), loadFolder()])
			form.theme = settings.theme
			form.root = settings.root
			form.port = settings.port
			form.hosts = settings.hosts
			form.developerToolsCLI = settings.developerToolsCLI
			listening.value = settings.listening
			nginxError.value = settings.error
			dataFolder.value = folder
			if (settings.theme === 'dark' || settings.theme === 'system') {
				setMode(settings.theme)
			} else {
				setMode('light')
			}
		} catch (err: unknown) {
			toast.error('读取设置失败', formatError(err))
		}
	}

	// 切换主题色彩模式并即时持久化
	async function selectTheme(mode: ThemeMode): Promise<void> {
		form.theme = mode
		setMode(mode)
		if (form.root !== '') {
			try {
				await saveSettings({
					theme: mode,
					root: form.root,
					port: Number(form.port),
					hosts: form.hosts,
					developerToolsCLI: form.developerToolsCLI,
				})
			} catch (err: unknown) {
				toast.error('保存主题失败', formatError(err))
			}
		}
	}

	// 切换主题强调色
	function selectColor(color: string): void {
		setColor(color)
	}

	// 选择默认项目父目录
	async function pickRoot(): Promise<void> {
		try {
			const picked = await pickFolder('选择默认项目父目录')
			if (picked !== '') {
				form.root = picked
			}
		} catch (err: unknown) {
			toast.error('选择目录失败', formatError(err))
		}
	}

	// 保存全部全局设置
	async function saveAll(): Promise<void> {
		saving.value = true
		try {
			const result = await saveSettings({
				theme: form.theme,
				root: form.root,
				port: Number(form.port),
				hosts: form.hosts,
				developerToolsCLI: form.developerToolsCLI,
			})
			if (form.theme === 'dark' || form.theme === 'system') {
				setMode(form.theme)
			} else {
				setMode('light')
			}
			await fetchData()
			if (!result.ok) {
				toast.error('保存设置存在问题', result.message, 4500)
				return
			}
			toast.success('设置已保存', result.message)
		} catch (err: unknown) {
			toast.error('保存设置失败', formatError(err))
		} finally {
			saving.value = false
		}
	}

	// 打开应用数据目录
	async function openData(): Promise<void> {
		if (dataFolder.value === '') {
			return
		}
		try {
			const result = await openFolder(dataFolder.value)
			if (!result.ok) {
				toast.error('打开目录失败', result.message)
			}
		} catch (err: unknown) {
			toast.error('打开目录失败', formatError(err))
		}
	}

	async function openNginxConfig(): Promise<void> {
		try {
			const result = await openNginxConfigApi()
			if (!result.ok) {
				toast.error('打开 Nginx 配置失败', result.message)
			}
		} catch (err: unknown) {
			toast.error('打开 Nginx 配置失败', formatError(err))
		}
	}

	async function pickDeveloperToolsCLI(): Promise<void> {
		try {
			const picked = await pickFile('选择微信开发者工具 CLI')
			if (picked !== '') {
				form.developerToolsCLI = picked
			}
		} catch (err: unknown) {
			toast.error('选择 CLI 失败', formatError(err))
		}
	}

	return {
		form,
		theme,
		THEME_COLORS,
		colorName,
		listening,
		nginxError,
		dataFolder,
		saving,
		fetchData,
		selectTheme,
		selectColor,
		pickRoot,
		saveAll,
		openData,
		openNginxConfig,
		pickDeveloperToolsCLI,
	}
}
