import type { App } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import { definePreset } from '@primevue/themes'
import Aura from '@primevue/themes/aura'
import { router } from '@/router'
import { pinia } from '@/store'
import { initTheme, setMode } from '@/composables/theme'
import { loadSettings } from '@/api'
import { toast, formatError } from '@/utils'

import 'primeicons/primeicons.css'
import '@/assets/css/style.css'
import '@/assets/css/primevue.css'

// PrimeVue Aura 主题预设
export const AuraPreset = definePreset(Aura, {
	semantic: {
		formField: {
			fontSize: 'var(--font-sm)',
			fontWeight: '400',
			sm: {
				fontSize: 'var(--font-sm)',
			},
			lg: {
				fontSize: 'var(--font-base)',
			},
		},
		primary: {
			50: '{slate.50}',
			100: '{slate.100}',
			200: '{slate.200}',
			300: '{slate.300}',
			400: '{slate.400}',
			500: '{slate.500}',
			600: '{slate.600}',
			700: '{slate.700}',
			800: '{slate.800}',
			900: '{slate.900}',
			950: '{slate.950}',
		},
		colorScheme: {
			light: {
				primary: {
					color: '{slate.700}',
					inverseColor: '#ffffff',
					hoverColor: '{slate.800}',
					activeColor: '{slate.900}',
				},
			},
			dark: {
				primary: {
					color: '{slate.300}',
					inverseColor: '{slate.950}',
					hoverColor: '{slate.200}',
					activeColor: '{slate.100}',
				},
			},
		},
	},
	components: {
		button: {
			root: {
				fontSize: 'var(--font-xs)',
				paddingX: '0.625rem',
				paddingY: '0.375rem',
				iconOnlyWidth: '2rem',
				sm: {
					fontSize: 'var(--font-xs)',
					paddingX: '0.5rem',
					paddingY: '0.25rem',
					iconOnlyWidth: '1.75rem',
				},
			},
		},
		message: {
			text: {
				fontSize: 'var(--font-sm)',
				fontWeight: '400',
				sm: { fontSize: 'var(--font-xs)' },
				lg: { fontSize: 'var(--font-base)' },
			},
			icon: {
				size: 'var(--icon-md)',
				sm: { size: 'var(--icon-sm)' },
				lg: { size: 'var(--icon-lg)' },
			},
		},
		panel: {
			title: { fontWeight: '400' },
		},
	},
})

// 启动引导注册插件与主题
export function bootstrap(app: App): void {
	app.use(pinia)
	app.use(router)
	app.use(PrimeVue, {
		ripple: false,
		theme: {
			preset: AuraPreset,
			options: {
				darkModeSelector: '.dark',
			},
		},
	})
	app.use(ToastService)
	app.use(ConfirmationService)

	initTheme()
	loadSettings()
		.then((settings) => {
			if (settings.theme === 'light' || settings.theme === 'dark' || settings.theme === 'system') {
				setMode(settings.theme)
			}
		})
		.catch((err: unknown) => {
			toast.error('加载设置失败', formatError(err))
		})
}
