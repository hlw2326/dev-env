import { ref, reactive, computed, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { loadSettings, saveSettings, pickFolder, openFolder, loadDataFolder } from '@/api'
import { useTheme, type ThemeMode } from '@/composables/theme'

export function useSettings() {
	const toast = useToast()
	const { theme, setThemeMode, setThemeColor, THEME_COLORS } = useTheme()

	const form = reactive({
		theme: 'light',
		root: '',
		port: 80,
		hosts: false,
		autostart: false,
	})
	const listening = ref(false)
	const gatewayError = ref('')
	const dataFolder = ref('')
	const saving = ref(false)

	const currentColorName = computed(() => {
		const matched = THEME_COLORS.find((item) => item.value.toLowerCase() === theme.color.toLowerCase())
		if (matched) {
			return matched.name
		}
		return theme.color
	})

	async function fetchSettings() {
		try {
			const [cfg, folder] = await Promise.all([loadSettings(), loadDataFolder()])
			form.theme = cfg.theme
			form.root = cfg.root
			form.port = cfg.port
			form.hosts = cfg.hosts
			form.autostart = cfg.autostart
			listening.value = cfg.listening
			gatewayError.value = cfg.error
			dataFolder.value = folder
			if (cfg.theme === 'dark' || cfg.theme === 'system') {
				setThemeMode(cfg.theme)
			} else {
				setThemeMode('light')
			}
		} catch (err) {
			toast.add({ severity: 'error', summary: '读取设置失败', detail: String(err), life: 3000 })
		}
	}

	async function selectTheme(mode: ThemeMode) {
		form.theme = mode
		setThemeMode(mode)
		if (form.root !== '') {
			await saveSettings({
				theme: mode,
				root: form.root,
				port: Number(form.port),
				hosts: form.hosts,
				autostart: form.autostart,
			})
		}
	}

	function selectColor(color: string) {
		setThemeColor(color)
	}

	async function pickRoot() {
		const picked = await pickFolder('选择默认站点父目录')
		if (picked !== '') {
			form.root = picked
		}
	}

	async function save() {
		saving.value = true
		try {
			const res = await saveSettings({
				theme: form.theme,
				root: form.root,
				port: Number(form.port),
				hosts: form.hosts,
				autostart: form.autostart,
			})
			if (form.theme === 'dark' || form.theme === 'system') {
				setThemeMode(form.theme)
			} else {
				setThemeMode('light')
			}
			await fetchSettings()
			if (!res.ok) {
				toast.add({ severity: 'error', summary: '保存设置存在问题', detail: res.message, life: 4500 })
				return
			}
			toast.add({ severity: 'success', summary: '设置已保存', detail: res.message, life: 2000 })
		} finally {
			saving.value = false
		}
	}

	async function openData() {
		if (dataFolder.value === '') {
			return
		}
		const res = await openFolder(dataFolder.value)
		if (!res.ok) {
			toast.add({ severity: 'error', summary: '打开目录失败', detail: res.message, life: 3000 })
		}
	}

	onMounted(() => {
		fetchSettings()
	})

	return {
		form,
		theme,
		THEME_COLORS,
		currentColorName,
		listening,
		gatewayError,
		dataFolder,
		saving,
		selectTheme,
		selectColor,
		pickRoot,
		save,
		openData,
	}
}
