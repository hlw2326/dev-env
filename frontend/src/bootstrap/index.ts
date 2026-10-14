import type { App } from 'vue'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import { definePreset } from '@primevue/themes'
import Aura from '@primevue/themes/aura'
import { router } from '@/router'
import { pinia } from '@/store'

import 'primeicons/primeicons.css'
import '@/style.css'

/**
 * 自定义 Slate 灰蓝极简浅色主题
 * 遵循 FRONTEND_STYLE_GUIDE.md 规范：主色调使用 Slate-700 (#334155)，小圆角，扁平化
 */
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
      950: '{slate.950}'
    },
    colorScheme: {
      light: {
        primary: {
          color: '{slate.700}',
          inverseColor: '#ffffff',
          hoverColor: '{slate.800}',
          activeColor: '{slate.900}'
        }
      }
    }
  }
})

export function bootstrap(app: App) {
  app.use(pinia)
  app.use(router)
  app.use(PrimeVue, {
    ripple: false, // 彻底禁用 PrimeVue 点击水波纹动效
    theme: {
      preset: DevEnvPreset,
      options: {
        darkModeSelector: false // 默认严格使用 Light 亮色主题
      }
    }
  })
  app.use(ToastService)
  app.use(ConfirmationService)
}
