<template>
  <h-page title="PHP 运行环境管理" desc="多版本 PHP 隔离共存、扩展模块配置与系统 CLI 默认版本控制">
    <template #extra>
      <p-button 
        label="登记本地 PHP" 
        icon="pi pi-plus" 
        @click="showRegisterModal = true" 
      />
    </template>

    <!-- Versions Grid -->
    <div v-if="versions.length === 0" class="surface-card empty-card">
      <span v-if="loading">加载运行环境中...</span>
      <span v-else>未找到已登记的 PHP 版本，点击上方按钮添加已有 PHP 目录</span>
    </div>

    <div v-else class="versions-grid">
      <div 
        v-for="ver in versions" 
        :key="ver.id" 
        class="version-card surface-card" 
        :class="{ 'is-default': ver.isDefault }"
      >
        <div class="card-top">
          <div class="php-badge">
            <span class="php-txt">PHP</span>
            <span class="php-v">{{ ver.version }}</span>
          </div>
          <p-tag v-if="ver.isDefault" severity="success" value="默认 CLI 版本" icon="pi pi-check" />
        </div>

        <div class="card-details">
          <div class="detail-row">
            <span class="detail-lbl">架构:</span>
            <span class="detail-val font-mono">{{ ver.architecture }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-lbl">安装目录:</span>
            <span class="detail-val font-mono path-truncate" :title="ver.installPath">{{ ver.installPath }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-lbl">状态:</span>
            <p-tag 
              :severity="ver.status === 'installed' ? 'success' : 'danger'"
              :value="ver.status === 'installed' ? '可运行' : '不可用'" 
              size="small"
            />
          </div>
        </div>

        <div class="card-actions">
          <p-button 
            label="扩展模块" 
            icon="pi pi-sliders-h" 
            severity="secondary" 
            outlined 
            size="small"
            @click="openExtensions(ver.id, ver.version)"
          />
          <p-button 
            v-if="!ver.isDefault" 
            label="设为默认" 
            size="small" 
            @click="handleSetDefault(ver.id)"
          />
        </div>
      </div>
    </div>

    <!-- Register Dialog -->
    <p-dialog 
      v-model:visible="showRegisterModal" 
      modal 
      header="登记本地 PHP 环境" 
      :style="{ width: '480px' }"
    >
      <form @submit.prevent="submitRegister">
        <div class="form-field">
          <label>安装目录绝对路径 (含 php.exe)</label>
          <p-input-text 
            v-model="registerForm.installPath" 
            required 
            placeholder="如 D:\phpstudy_pro\Extensions\php\php8.2.9nts" 
          />
        </div>

        <div class="form-field">
          <label>CPU 架构</label>
          <p-select 
            v-model="registerForm.architecture" 
            :options="archOptions" 
            optionLabel="label" 
            optionValue="value" 
          />
        </div>

        <div class="dialog-footer">
          <p-button label="取消" severity="secondary" text @click="showRegisterModal = false" />
          <p-button label="确认登记" type="submit" :loading="submitting" />
        </div>
      </form>
    </p-dialog>

    <!-- Extensions Dialog -->
    <p-dialog 
      v-model:visible="showExtModal" 
      modal 
      :header="'PHP ' + currentExtVersion + ' 扩展模块清单'" 
      :style="{ width: '600px', maxHeight: '80vh' }"
    >
      <p class="ext-desc">从 php.exe 检测到的当前生效与可用扩展模块</p>

      <div v-if="extLoading" class="empty-state">读取扩展信息中...</div>
      <div v-else class="extensions-grid">
        <div v-for="ext in extensions" :key="ext.name" class="ext-badge-card">
          <span class="ext-title font-mono">{{ ext.name }}</span>
          <p-tag 
            :severity="ext.enabled ? 'success' : 'secondary'"
            :value="ext.enabled ? '已启用' : '可用'" 
            size="small"
          />
        </div>
      </div>

      <div class="dialog-footer">
        <p-button label="关闭" severity="secondary" text @click="showExtModal = false" />
      </div>
    </p-dialog>
  </h-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import HPage from '@/components/h-page/h-page.vue'
import { listVersions, registerVersion, setDefaultVersion, getExtensions } from '@/api/php'
import type { PHPVersion, PHPExtension } from '@/types'

defineOptions({
  name: 'PHP'
})

const versions = ref<PHPVersion[]>([])
const loading = ref(false)
const submitting = ref(false)

const showRegisterModal = ref(false)
const registerForm = ref({
  installPath: '',
  architecture: 'x64',
})

const archOptions = [
  { label: 'x64 (64位架构)', value: 'x64' },
  { label: 'x86 (32位架构)', value: 'x86' },
]

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
    toast.success('PHP 版本已成功登记')
    showRegisterModal.value = false
    registerForm.value.installPath = ''
    await loadVersions()
  } catch (err: any) {
    toast.error(err?.message || '登记失败')
  } finally {
    submitting.value = false
  }
}

async function handleSetDefault(id: number) {
  try {
    await setDefaultVersion(id)
    toast.success('已设置为默认 CLI 版本')
    await loadVersions()
  } catch (err: any) {
    toast.error(err?.message || '设置默认失败')
  }
}

async function openExtensions(id: number, verName: string) {
  currentExtVersion.value = verName
  showExtModal.value = true
  extLoading.value = true
  try {
    extensions.value = await getExtensions(id)
  } catch (err: any) {
    toast.error(err?.message || '获取扩展信息失败')
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
}

.versions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 16px;
}

.version-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.version-card.is-default {
  border-color: #93c5fd;
  box-shadow: 0 0 0 1px #bfdbfe;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.php-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.php-txt {
  background: #3b82f6;
  color: #ffffff;
  font-weight: 400;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.php-v {
  font-size: 1.25rem;
  font-weight: 400;
  color: var(--text-main);
}

.card-details {
  background: #f8fafc;
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
}

.detail-lbl {
  color: var(--text-muted);
}

.detail-val {
  color: var(--text-main);
  font-weight: 400;
}

.path-truncate {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.ext-desc {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: -6px;
  margin-bottom: 14px;
}

.extensions-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  max-height: 420px;
  overflow-y: auto;
  padding: 4px;
}

.ext-badge-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-sm);
}

.ext-title {
  font-size: 0.8rem;
  color: var(--text-main);
}

.empty-card {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}
</style>
