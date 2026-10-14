<template>
  <h-page :title="$t('settings.title')" :desc="$t('settings.desc')">
    <template #extra>
      <p-button :label="$t('settings.reset')" severity="secondary" outlined @click="handleReset" />
      <p-button :label="$t('settings.save')" icon="pi pi-check" :loading="saving" @click="handleSave" />
    </template>

    <div class="settings-grid">
      <!-- 通用外观与语言 -->
      <h-card :title="$t('settings.appearance')" class="full-col">
        <!-- 界面语言选择 -->
        <div class="setting-item">
          <label class="setting-item-label">{{ $t('settings.lang') }}</label>
          <div class="lang-switch-group">
            <button
              type="button"
              class="lang-pill"
              :class="{ active: currentLocale === 'zh-CN' }"
              @click="handleLanguageChange('zh-CN')"
            >
              <i class="pi pi-globe"></i>
              <span>简体中文</span>
            </button>
            <button
              type="button"
              class="lang-pill"
              :class="{ active: currentLocale === 'en-US' }"
              @click="handleLanguageChange('en-US')"
            >
              <i class="pi pi-language"></i>
              <span>English</span>
            </button>
          </div>
        </div>

        <div class="divider"></div>

        <!-- 色彩模式 -->
        <div class="setting-item">
          <label class="setting-item-label">{{ $t('settings.themeMode') }}</label>
          <div class="theme-options">
            <!-- 浅色明亮 -->
            <div
              class="theme-card"
              :class="{ active: theme.mode === 'light' }"
              @click="handleThemeModeChange('light')"
            >
              <div class="theme-preview light-preview">
                <div class="bar"></div>
                <div class="line"></div>
              </div>
              <span class="theme-name">{{ $t('settings.light') }}</span>
            </div>

            <!-- 深色暗黑 -->
            <div
              class="theme-card"
              :class="{ active: theme.mode === 'dark' }"
              @click="handleThemeModeChange('dark')"
            >
              <div class="theme-preview dark-preview">
                <div class="bar"></div>
                <div class="line"></div>
              </div>
              <span class="theme-name">{{ $t('settings.dark') }}</span>
            </div>

            <!-- 跟随系统 -->
            <div
              class="theme-card"
              :class="{ active: theme.mode === 'system' }"
              @click="handleThemeModeChange('system')"
            >
              <div class="theme-preview system-preview">
                <div class="half-light"></div>
                <div class="half-dark"></div>
              </div>
              <span class="theme-name">{{ $t('settings.system') }}</span>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <!-- 系统强调色 -->
        <div class="setting-item">
          <div class="color-header">
            <label class="setting-item-label">{{ $t('settings.themeColor') }}</label>
            <span class="color-tag" :style="{ backgroundColor: theme.color }">
              {{ currentColorName }}
            </span>
          </div>

          <div class="color-options">
            <div
              v-for="item in THEME_COLORS"
              :key="item.value"
              class="color-wrap"
              :class="{ active: theme.color.toLowerCase() === item.value.toLowerCase() }"
              :title="item.name"
              @click="setThemeColor(item.value)"
            >
              <div class="color-dot" :style="{ backgroundColor: item.value }">
                <i
                  v-if="theme.color.toLowerCase() === item.value.toLowerCase()"
                  class="pi pi-check check-icon"
                ></i>
              </div>
            </div>

            <!-- 自定义颜色拾取器 -->
            <label class="color-btn" :title="$t('settings.customColor')">
              <input
                type="color"
                :value="theme.color"
                class="color-input"
                @input="onColorPick"
              />
              <i class="pi pi-plus plus-icon"></i>
            </label>
          </div>
        </div>
      </h-card>

      <!-- 核心工作路径 -->
      <h-card :title="$t('settings.paths')">
        <div class="form-field">
          <label>{{ $t('settings.runtimeRoot') }}</label>
          <p-input-text
            v-model="settings.runtimeRootPath"
            class="font-mono"
            placeholder="如 D:\DevEnv\runtimes"
          />
        </div>

        <div class="form-field">
          <label>{{ $t('settings.defaultProject') }}</label>
          <p-input-text
            v-model="settings.defaultProjectPath"
            class="font-mono"
            placeholder="如 G:\www 或 D:\projects"
          />
        </div>
      </h-card>

      <!-- 服务管理 -->
      <h-card :title="$t('settings.serviceManager')">
        <div class="toggle-row">
          <div class="toggle-info">
            <span class="toggle-title">{{ $t('settings.autoStart') }}</span>
            <span class="toggle-sub">{{ $t('settings.autoStartSub') }}</span>
          </div>
          <p-toggle-switch v-model="settings.autoStartServices" />
        </div>
      </h-card>
    </div>
  </h-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import HPage from '@/components/h-page/h-page.vue'
import HCard from '@/components/h-card/h-card.vue'
import { getSettings, updateSettings, resetSettings } from '@/api/settings'
import { $t, setLocale, getLocale } from '@/i18n'
import { useTheme, type ThemeMode } from '@/composables/theme'
import type { AppSettings } from '@/types'

defineOptions({
  name: 'Settings'
})

const { theme, setThemeMode, setThemeColor, THEME_COLORS } = useTheme()

const currentLocale = computed(() => getLocale())

const currentColorName = computed(() => {
  const matched = THEME_COLORS.find(
    (c) => c.value.toLowerCase() === theme.color.toLowerCase()
  )
  return matched ? matched.name : `${$t('settings.customColor')} (${theme.color})`
})

const settings = ref<AppSettings>({
  language: 'zh-CN',
  theme: 'light',
  runtimeRootPath: '',
  defaultProjectPath: '',
  autoStartServices: false
})

const saving = ref(false)

function handleLanguageChange(lang: string) {
  settings.value.language = lang
  setLocale(lang)
}

function handleThemeModeChange(mode: ThemeMode) {
  settings.value.theme = mode
  setThemeMode(mode)
}

function onColorPick(event: Event) {
  const color = (event.target as HTMLInputElement).value
  setThemeColor(color)
}

async function loadSettings() {
  try {
    const data = await getSettings()
    settings.value = data
    if (data.language) setLocale(data.language)
    if (data.theme) setThemeMode(data.theme)
  } catch (err: any) {
    console.error('加载设置失败', err)
  }
}

async function handleSave() {
  saving.value = true
  try {
    settings.value.theme = theme.mode
    settings.value.language = getLocale()
    settings.value = await updateSettings(settings.value)
    toast.success($t('settings.saveSuccess'))
  } catch (err: any) {
    toast.error(err?.message || '保存设置失败')
  } finally {
    saving.value = false
  }
}

async function handleReset() {
  if (!confirm($t('settings.resetConfirm'))) return
  try {
    settings.value = await resetSettings()
    setLocale(settings.value.language || 'zh-CN')
    setThemeMode((settings.value.theme as ThemeMode) || 'light')
    setThemeColor('#334155')
    toast.info($t('settings.resetSuccess'))
  } catch (err: any) {
    toast.error(err?.message || '重置失败')
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped lang="scss">
.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--h-padding);
}

.full-col {
  grid-column: span 2;
}

.setting-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.setting-item-label {
  font-size: var(--font-sm);
  font-weight: 500;
  color: var(--text-primary);
}

.divider {
  height: 1px;
  background: var(--border-color);
  margin: 12px 0;
}

/* 语言切换药丸组 */
.lang-switch-group {
  display: flex;
  gap: 10px;

  .lang-pill {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 18px;
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color);
    background: var(--bg-card);
    color: var(--text-secondary);
    font-size: var(--font-sm);
    cursor: pointer;
    transition: all 0.15s ease;

    &:hover {
      border-color: var(--p-primary-color);
      color: var(--text-primary);
    }

    &.active {
      border-color: var(--p-primary-color);
      background: var(--bg-app);
      color: var(--p-primary-color);
      font-weight: 500;
    }
  }
}

/* 主题模式微缩卡片（参考 zbzs/ele） */
.theme-options {
  display: flex;
  gap: 1.25rem;
}

.theme-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;

  .theme-preview {
    width: 6.5rem;
    height: 4rem;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-color);
    position: relative;
    overflow: hidden;
    transition: all 0.18s ease;

    .bar {
      height: 0.875rem;
      width: 100%;
    }

    .line {
      height: 0.375rem;
      width: 60%;
      margin: 0.75rem auto;
      border-radius: 2px;
    }
  }

  .light-preview {
    background: #ffffff;
    .bar { background: #f1f5f9; }
    .line { background: #e2e8f0; }
  }

  .dark-preview {
    background: #0f172a;
    .bar { background: #1e293b; }
    .line { background: #334155; }
  }

  .system-preview {
    display: flex;
    .half-light { width: 50%; height: 100%; background: #ffffff; }
    .half-dark { width: 50%; height: 100%; background: #0f172a; }
  }

  .theme-name {
    font-size: var(--font-xs);
    color: var(--text-secondary);
    transition: color 0.18s ease;
  }

  &:hover {
    .theme-preview {
      border-color: var(--p-primary-color);
    }
  }

  &.active {
    .theme-preview {
      border-color: var(--p-primary-color);
      box-shadow: 0 0 0 2px var(--p-primary-color);
    }

    .theme-name {
      color: var(--p-primary-color);
      font-weight: 500;
    }
  }
}

/* 系统强调色选择 */
.color-header {
  display: flex;
  align-items: center;
  gap: 10px;

  .color-tag {
    font-size: var(--font-xs);
    color: #ffffff;
    padding: 2px 8px;
    border-radius: 4px;
    line-height: 1.2;
    font-weight: 500;
  }
}

.color-options {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;

  .color-wrap {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    cursor: pointer;
    border: 2px solid transparent;
    transition: all 0.15s ease;

    &:hover {
      transform: scale(1.1);
    }

    &.active {
      border-color: var(--text-primary);
    }

    .color-dot {
      width: 1.375rem;
      height: 1.375rem;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;

      .check-icon {
        font-size: 0.75rem;
        color: #ffffff;
      }
    }
  }

  .color-btn {
    width: 1.875rem;
    height: 1.875rem;
    border-radius: 50%;
    border: 1px dashed var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    position: relative;
    transition: all 0.15s ease;

    &:hover {
      border-color: var(--p-primary-color);
      color: var(--p-primary-color);
    }

    .color-input {
      position: absolute;
      opacity: 0;
      width: 100%;
      height: 100%;
      cursor: pointer;
    }

    .plus-icon {
      font-size: 0.75rem;
      color: var(--text-secondary);
    }
  }
}

.toggle-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toggle-title {
  font-size: var(--font-base);
  color: var(--text-primary);
}

.toggle-sub {
  font-size: var(--font-xs);
  color: var(--text-muted);
}
</style>
