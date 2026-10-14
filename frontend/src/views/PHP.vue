<template>
  <div class="php-page fade-in">
    <!-- Header -->
    <div class="page-header glass-panel">
      <div class="header-left">
        <h2>PHP 运行环境管理</h2>
        <p>多版本 PHP 隔离共存、扩展模块配置与系统 CLI 默认版本控制</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="showRegisterModal = true">
          <PlusIcon class="btn-icon" />
          登记本地 PHP
        </button>
      </div>
    </div>

    <!-- PHP Versions Cards Grid -->
    <div v-if="versions.length === 0" class="glass-panel empty-state">
      <span v-if="loading">加载运行环境中...</span>
      <span v-else>未找到已登记的 PHP 版本，点击上方按钮添加已有 PHP 目录</span>
    </div>

    <div v-else class="versions-grid">
      <div v-for="ver in versions" :key="ver.id" class="version-card glass-panel" :class="{ 'is-default': ver.isDefault }">
        <div class="card-header">
          <div class="ver-brand">
            <span class="php-tag">PHP</span>
            <span class="ver-num">{{ ver.version }}</span>
          </div>
          <span v-if="ver.isDefault" class="badge badge-success">默认 CLI 版本</span>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="info-label">CPU 架构:</span>
            <span class="info-val font-mono">{{ ver.architecture }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">安装目录:</span>
            <span class="info-val font-mono path-text" :title="ver.installPath">{{ ver.installPath }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">运行状态:</span>
            <span class="badge" :class="ver.status === 'installed' ? 'badge-success' : 'badge-danger'">
              {{ ver.status === 'installed' ? '可运行' : '不可用' }}
            </span>
          </div>
        </div>

        <div class="card-footer">
          <button class="btn btn-secondary btn-sm" @click="openExtensions(ver.id, ver.version)">
            <SlidersIcon class="btn-icon" />
            扩展模块 (Extensions)
          </button>
          <button 
            v-if="!ver.isDefault" 
            class="btn btn-primary btn-sm" 
            @click="handleSetDefault(ver.id)"
          >
            设为默认
          </button>
        </div>
      </div>
    </div>

    <!-- Register Modal -->
    <div v-if="showRegisterModal" class="modal-overlay" @click.self="showRegisterModal = false">
      <div class="modal-content">
        <h3>登记本地 PHP 环境</h3>
        <p class="modal-subtitle">指定含有 php.exe 的本地目录进行快速注册</p>

        <form @submit.prevent="submitRegister">
          <div class="form-group">
            <label class="form-label">安装目录路径</label>
            <input 
              v-model="registerForm.installPath" 
              type="text" 
              class="form-input" 
              required 
              placeholder="如 D:\phpstudy_pro\Extensions\php\php8.2.9nts" 
            />
          </div>

          <div class="form-group">
            <label class="form-label">架构</label>
            <select v-model="registerForm.architecture" class="form-input">
              <option value="x64">x64 (64位)</option>
              <option value="x86">x86 (32位)</option>
            </select>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showRegisterModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">确认登记</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Extensions Modal -->
    <div v-if="showExtModal" class="modal-overlay" @click.self="showExtModal = false">
      <div class="modal-content ext-modal-content">
        <div class="ext-modal-header">
          <h3>PHP {{ currentExtVersion }} 扩展模块</h3>
          <button class="btn btn-secondary btn-sm" @click="showExtModal = false">关闭</button>
        </div>
        <p class="modal-subtitle">从 php.exe 检测到的当前生效与可用扩展清单</p>

        <div v-if="extLoading" class="empty-state">读取扩展中...</div>
        <div v-else class="extensions-scroll">
          <div v-for="ext in extensions" :key="ext.name" class="ext-item">
            <span class="ext-name font-mono">{{ ext.name }}</span>
            <span class="badge" :class="ext.enabled ? 'badge-success' : 'badge-neutral'">
              {{ ext.enabled ? '已启用 (Active)' : '可用 (Available)' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus as PlusIcon, Sliders as SlidersIcon } from 'lucide-vue-next'
import { listVersions, registerVersion, setDefaultVersion, getExtensions } from '../api/php'
import type { PHPVersion, PHPExtension } from '../types'

const versions = ref<PHPVersion[]>([])
const loading = ref(false)
const submitting = ref(false)

const showRegisterModal = ref(false)
const registerForm = ref({
  installPath: '',
  architecture: 'x64',
})

const showExtModal = ref(false)
const extLoading = ref(false)
const currentExtVersion = ref('')
const extensions = ref<PHPExtension[]>([])

async function loadVersions() {
  loading.value = true
  try {
    versions.value = await listVersions()
  } finally {
    loading.value = false
  }
}

async function submitRegister() {
  submitting.value = true
  try {
    await registerVersion({
      version: 'Auto',
      architecture: registerForm.value.architecture,
      installPath: registerForm.value.installPath,
    })
    showRegisterModal.value = false
    registerForm.value.installPath = ''
    await loadVersions()
  } catch (err: any) {
    alert(err?.message || '登记失败')
  } finally {
    submitting.value = false
  }
}

async function handleSetDefault(id: number) {
  try {
    await setDefaultVersion(id)
    await loadVersions()
  } catch (err: any) {
    alert(err?.message || '设置默认失败')
  }
}

async function openExtensions(id: number, verName: string) {
  currentExtVersion.value = verName
  showExtModal.value = true
  extLoading.value = true
  try {
    extensions.value = await getExtensions(id)
  } catch (err: any) {
    alert(err?.message || '获取扩展信息失败')
  } finally {
    extLoading.value = false
  }
}

onMounted(() => {
  loadVersions()
})
</script>

<style scoped>
.php-page {
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

.versions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.version-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  border-radius: var(--radius-lg);
  transition: all 0.2s;
}
.version-card.is-default {
  border-color: rgba(99, 102, 241, 0.4);
  box-shadow: 0 0 16px rgba(99, 102, 241, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ver-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.php-tag {
  background: #6366f1;
  color: #fff;
  font-weight: 700;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.ver-num {
  font-size: 1.25rem;
  font-weight: 700;
  color: #fff;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: rgba(15, 23, 42, 0.4);
  padding: 12px;
  border-radius: var(--radius-md);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
}

.info-label {
  color: var(--text-dim);
}

.info-val {
  color: var(--text-main);
}

.path-text {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.ext-modal-content {
  max-width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.ext-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.extensions-scroll {
  overflow-y: auto;
  max-height: 400px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 12px;
}

.ext-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.ext-name {
  font-size: 0.8rem;
  color: var(--text-main);
}

.font-mono {
  font-family: var(--font-mono);
}

.empty-state {
  text-align: center;
  color: var(--text-dim);
  padding: 40px;
  font-size: 0.875rem;
}
</style>
