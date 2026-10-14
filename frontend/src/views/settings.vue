<template>
  <h-page title="应用全局设置" desc="配置运行环境根目录、默认项目存储路径、启动偏好与内部数据库">
    <template #extra>
      <p-button label="恢复默认" severity="secondary" outlined @click="handleReset" />
      <p-button label="保存配置" icon="pi pi-check" :loading="saving" @click="handleSave" />
    </template>

    <!-- Settings Grid -->
    <div class="settings-grid">
      <!-- General -->
      <h-card title="通用外观与语言">

        <div class="form-field">
          <label>界面语言</label>
          <p-select 
            v-model="settings.language" 
            :options="langOptions" 
            optionLabel="label" 
            optionValue="value" 
          />
        </div>

        <div class="form-field">
          <label>主题模式</label>
          <p-select 
            v-model="settings.theme" 
            :options="themeOptions" 
            optionLabel="label" 
            optionValue="value" 
          />
        </div>
      </h-card>

      <!-- Environment Paths -->
      <h-card title="路径配置">
        <div class="form-field">
          <label>环境根目录</label>
          <p-input-text 
            v-model="settings.runtimeRootPath" 
            class="font-mono" 
            placeholder="如 D:\DevEnv\runtimes" 
          />
        </div>

        <div class="form-field">
          <label>默认项目路径</label>
          <p-input-text 
            v-model="settings.defaultProjectPath" 
            class="font-mono" 
            placeholder="如 G:\www 或 D:\projects" 
          />
        </div>
      </h-card>

      <!-- Services & DB Info -->
      <h-card title="服务管理与内部数据持久化" class="full-col">
        <div class="toggle-row">
          <div class="toggle-info">
            <span class="toggle-title">开机自启服务</span>
            <span class="toggle-sub">应用启动时自动开启标记为自启的服务实例</span>
          </div>
          <p-toggle-switch v-model="settings.autoStartServices" />
        </div>

        <div class="db-spec-box font-mono">
          <div class="spec-row">
            <span class="spec-lbl">数据库驱动:</span>
            <span class="spec-val">SQLite 3 (WAL 模式, 纯 Go 驱动 modernc.org/sqlite)</span>
          </div>
          <div class="spec-row">
            <span class="spec-lbl">统一表名前缀:</span>
            <span class="spec-val">system_projects, system_runtime_versions, system_service_instances, system_virtual_hosts, system_tasks, system_app_settings</span>
          </div>
          <div class="spec-row">
            <span class="spec-lbl">主键设计规范:</span>
            <span class="spec-val">数字自增 ID (INTEGER PRIMARY KEY AUTOINCREMENT)</span>
          </div>
        </div>
      </h-card>
    </div>
  </h-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import HPage from '@/components/h-page/h-page.vue'
import HCard from '@/components/h-card/h-card.vue'
import { getSettings, updateSettings, resetSettings } from '@/api/settings'
import type { AppSettings } from '@/types'

defineOptions({
  name: 'Settings'
})

const settings = ref<AppSettings>({
  language: 'zh-CN',
  theme: 'light',
  runtimeRootPath: '',
  defaultProjectPath: '',
  autoStartServices: false,
})

const saving = ref(false)

const langOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' },
]

const themeOptions = [
  { label: '明亮浅色', value: 'light' },
  { label: '跟随系统', value: 'system' },
  { label: '暗黑深色', value: 'dark' },
]

async function loadSettings() {
  try {
    settings.value = await getSettings()
  } catch (err: any) {
    console.error('加载设置失败', err)
  }
}

async function handleSave() {
  saving.value = true
  try {
    settings.value = await updateSettings(settings.value)
    toast.success('设置已成功保存')
  } catch (err: any) {
    toast.error(err?.message || '保存设置失败')
  } finally {
    saving.value = false
  }
}

async function handleReset() {
  if (!confirm('确定要恢复默认设置吗？已有项目和环境记录不受影响。')) return
  try {
    settings.value = await resetSettings()
    toast.info('已重置为默认配置')
  } catch (err: any) {
    toast.error(err?.message || '重置失败')
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
}

.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.setting-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.setting-card.full-col {
  grid-column: span 2;
}

.card-title {
  font-size: 1.05rem;
  font-weight: 400;
  color: var(--text-main);
}

.card-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: -10px;
}

.toggle-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  background: #f8fafc;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.toggle-info {
  display: flex;
  flex-direction: column;
}

.toggle-title {
  font-size: 0.9rem;
  font-weight: 400;
  color: var(--text-main);
}

.toggle-sub {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.db-spec-box {
  background: #f8fafc;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 0.775rem;
}

.spec-row {
  display: flex;
  gap: 12px;
}

.spec-lbl {
  color: var(--text-muted);
  min-width: 120px;
}

.spec-val {
  color: #2563eb;
}
</style>
