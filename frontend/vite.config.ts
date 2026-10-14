import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { PrimeVueResolver } from '@primevue/auto-import-resolver'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      dirs: ['src/utils'],
      dts: 'src/types/auto-imports.d.ts'
    }),
    Components({
      resolvers: [
        PrimeVueResolver(),
        PrimeVueResolver({ components: { prefix: 'P' } }),
        PrimeVueResolver({ components: { prefix: 'p' } })
      ],
      dirs: ['src/components'],
      dts: 'src/types/components.d.ts'
    })
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
})
