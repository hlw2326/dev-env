import { ref, onMounted } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import { getSettings, updateSettings, resetSettings, pickDir } from '@/api/settings'
import { $t, setLocale, getLocale } from '@/i18n'
import { useTheme, DEFAULT_THEME, type ThemeMode } from '@/composables/theme'
import type { AppSettings } from '@/types'

// 主题控制实例
const { theme, setThemeMode, setThemeColor } = useTheme()

// 系统设置响应式对象
const settings = ref<AppSettings>({
	language: 'zh-CN',
	theme: 'light',
	runtimeRootPath: '',
	defaultProjectPath: '',
	autoStartServices: false
})

// 保存按钮加载状态
const saving = ref(false)

// 选择运行环境安装根目录
async function pickRuntimeDir() {
	const dir = await pickDir('选择运行环境安装根目录')
	if (dir) {
		settings.value.runtimeRootPath = dir
	}
}

// 选择默认项目根目录
async function pickProjectDir() {
	const dir = await pickDir('选择默认网站/项目根目录')
	if (dir) {
		settings.value.defaultProjectPath = dir
	}
}

// 切换界面语言
function changeLang(lang: string) {
	settings.value.language = lang
	setLocale(lang)
}

// 切换主题色彩模式
function changeMode(mode: ThemeMode) {
	settings.value.theme = mode
	setThemeMode(mode)
}

// 加载系统设置
async function loadSettings() {
	try {
		const data = await getSettings()
		settings.value = data
		if (data.language) {
			setLocale(data.language)
		}
		if (data.theme) {
			setThemeMode(data.theme)
		}
	} catch (err) {
		console.error('加载设置失败', err)
	}
}

// 保存系统设置
async function saveConfig() {
	saving.value = true
	try {
		settings.value.theme = theme.mode
		settings.value.language = getLocale()
		settings.value = await updateSettings(settings.value)
		toast.success($t('设置已成功保存'))
	} catch (err) {
		toast.error(String(err))
	} finally {
		saving.value = false
	}
}

// 导出设置页逻辑
export function useSettings() {
	// 气泡确认框实例
	const confirm = useConfirm()

	// 重置系统设置
	function resetConfig(event: Event) {
		confirm.require({
			target: event.currentTarget as HTMLElement,
			message: $t('确定要将所有设置恢复为默认值吗？'),
			icon: 'pi pi-exclamation-triangle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: $t('恢复默认'),
				severity: 'danger',
				size: 'small'
			},
			accept: async () => {
				try {
					settings.value = await resetSettings()
					setLocale(settings.value.language)
					setThemeMode(settings.value.theme as ThemeMode)
					setThemeColor(DEFAULT_THEME.color)
					toast.info($t('已恢复默认设置'))
				} catch (err) {
					toast.error(String(err))
				}
			}
		})
	}

	onMounted(loadSettings)
	return {
		settings,
		saving,
		pickRuntimeDir,
		pickProjectDir,
		changeLang,
		changeMode,
		saveConfig,
		resetConfig
	}
}
