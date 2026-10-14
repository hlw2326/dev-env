<template>
  <div class="settings-page fade-in">
    <!-- Header -->
    <div class="page-header glass-panel">
      <div class="header-left">
        <h2>应用全局设置</h2>
        <p>配置运行环境根目录、默认项目存储路径、启动偏好与内部数据库</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="handleReset">恢复默认</button>
        <button class="btn btn-primary" :disabled="saving" @click="handleSave">保存配置</button>
      </div>
    </div>

    <!-- Settings Sections -->
    <div class="settings-grid">
      <!-- General Preferences -->
      <div class="settings-card glass-panel">
        <h3>通用偏好</h3>
        <p class="section-desc">界面语言与视觉外观设置</p>

        <div class="form-group">
          <label class="form-label">界面语言</label>
          <select v-model="settings.language" class="form-input">
            <option value="zh-CN">简体中文 (zh-CN)</option>
            <option value="en-US">English (en-US)</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">主题风格</label>
          <select v-model="settings.theme" class="form-input">
            <option value="system">跟随系统 (System)</option>
            <option value="dark">深色黑金 (Dark Slate)</option>
            <option value="light">浅色明亮 (Light)</option>
          </select>
        </div>
      </div>

      <!-- Environment Paths -->
      <div class="settings-card glass-panel">
        <h3>环境与项目目录</h3>
        <p class="section-desc">运行环境下载解压根目录与默认项目路径</p>

        <div class="form-group">
          <label class="form-label">环境根目录 (Runtime Root)</label>
          <input 
            v-model="settings.runtimeRootPath" 
            type="text" 
            class="form-input font-mono" 
            placeholder="如 D:\DevEnv\runtimes" 
          />
        </div>

        <div class="form-group">
          <label class="form-label">默认项目创建路径 (Projects Default)</label>
          <input 
            v-model="settings.defaultProjectPath" 
            type="text" 
            class="form-input font-mono" 
            placeholder="如 G:\www 或 D:\projects" 
          />
        </div>
      </div>

      <!-- Service Startup & System DB Info -->
      <div class="settings-card glass-panel full-width">
        <h3>服务与内部数据库 (SQLite)</h3>
        <p class="section-desc">应用底层管理数据库与外部服务生命周期策略</p>

        <div class="setting-checkbox-row">
          <label class="switch-label">
            <input v-model="settings.autoStartServices" type="checkbox" />
            <span>应用启动时自动开启标记为开机自启的服务实例</span>
          </label>
        </div>

        <div class="db-info-box font-mono">
          <div class="db-row">
            <span class="db-label">数据库引擎:</span>
            <span class="db-val">SQLite 3 (WAL 模式, 纯 Go 驱动 modernc.org/sqlite)</span>
          </div>
          <div class="db-row">
            <span class="db-label">存储数据表:</span>
            <span class="db-val">system_projects, system_runtime_versions, system_service_instances, system_virtual_hosts, system_tasks, system_app_settings</span>
          </div>
          <div class="db-row">
            <span class="db-label">ID 设计规范:</span>
            <span class="db-val">数字自增 ID (INTEGER PRIMARY KEY AUTOINCREMENT)</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSettings, updateSettings, resetSettings } from '../api/settings'
import type { AppSettings } from '../types'

const settings = ref<AppSettings>({
  language: 'zh-CN',
  theme: 'system',
  runtimeRootPath: '',
  defaultProjectPath: '',
  autoStartServices: false,
})

const saving = ref(false)

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
    alert('设置已成功保存！')
  } catch (err: any) {
    alert(err?.message || '保存设置失败')
  } finally {
    saving.value = false
  }
}

async function handleReset() {
  if (!confirm('确定要恢复默认设置吗？已有的项目和运行环境不会被删除。')) return
  try {
    settings.value = await resetSettings()
    alert('已恢复为默认配置')
  } catch (err: any) {
    alert(err?.message || '恢复设置失败')
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
  gap: 20px;
}

.page-header {
  padding: 20px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left h2 {
  font-size: 1.4rem;
  font-weight: 700;
  color: #fff;
}
.header-left p {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 3px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.settings-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.settings-card.full-width {
  grid-column: span 2;
}

.settings-card h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #fff;
}

.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: -10px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.setting-checkbox-row {
  display: flex;
  align-items: center;
}

.switch-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.875rem;
  color: var(--text-main);
  cursor: pointer;
}

.switch-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: var(--accent-primary);
  cursor: pointer;
}

.db-info-box {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 0.775rem;
}

.db-row {
  display: flex;
  gap: 12px;
}

.db-label {
  color: var(--text-dim);
  min-width: 120px;
}

.db-val {
  color: var(--accent-cyan);
}

.font-mono {
  font-family: var(--font-mono);
}
</style>
