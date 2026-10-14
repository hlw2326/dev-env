import type { App } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import { definePreset } from '@primevue/themes'
import Aura from '@primevue/themes/aura'
import { router } from '@/router'
import { pinia } from '@/store'
import { initTheme, setThemeMode } from '@/composables/theme'
import { loadSettings } from '@/api'

import 'primeicons/primeicons.css'
import '@/assets/css/style.css'
import '@/assets/css/primevue.css'

export const DevEnvPreset = definePreset(Aura, {
	semantic: {
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
})

export function bootstrap(app: App) {
	app.use(pinia)
	app.use(router)
	app.use(PrimeVue, {
		ripple: false,
		theme: {
			preset: DevEnvPreset,
			options: {
				darkModeSelector: '.dark',
			},
		},
	})
	app.use(ToastService)
	app.use(ConfirmationService)

	initTheme()
	loadSettings()
		.then((cfg) => {
			if (cfg.theme === 'light' || cfg.theme === 'dark' || cfg.theme === 'system') {
				setThemeMode(cfg.theme)
			}
		})
		.catch(() => {})
}
