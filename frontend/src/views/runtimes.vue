<template>
  <h-page title="运行环境" desc="统一管理已安装的 PHP、Nginx、MySQL 等本地开发环境版本">
    <template #extra>
      <p-button 
        label="刷新扫描" 
        icon="pi pi-refresh" 
        severity="secondary" 
        outlined 
        :loading="loading" 
        @click="loadRuntimes" 
      />
      <p-button 
        label="软件商店" 
        icon="pi pi-shopping-bag" 
        @click="$router.push('/store')" 
      />
    </template>

    <!-- 运行环境列表 (按类型分组) -->
    <div v-if="allRuntimes.length === 0 && !loading" class="empty-runtimes">
      <i class="pi pi-box empty-icon"></i>
      <p class="empty-title">暂未安装任何运行环境</p>
      <p class="empty-subtitle">前往软件商店可一键快速下载并安装 PHP、Nginx 等开发环境</p>
      <p-button 
        label="前往软件商店安装" 
        icon="pi pi-arrow-right" 
        @click="$router.push('/store')" 
        class="mt-2"
      />
    </div>

    <div v-else class="groups-container">
      <div 
        v-for="(items, type) in groupedRuntimes" 
        :key="type" 
        class="runtime-group"
      >
        <!-- 分组折叠头部 -->
        <div class="group-header" @click="toggleGroup(type)">
          <div class="group-header-left">
            <i :class="isGroupExpanded(type) ? 'pi pi-chevron-down' : 'pi pi-chevron-right'"></i>
            <div class="group-icon-badge" :style="{ backgroundColor: getRuntimeBrandColor(type) }">
              <i :class="getRuntimeIcon(type)"></i>
            </div>
            <span class="group-name">{{ getRuntimeTypeName(type) }}</span>
            <p-tag :value="`${items.length} 个版本`" severity="secondary" class="count-tag" />
          </div>
          <div class="group-header-right">
            <span class="collapse-hint">{{ isGroupExpanded(type) ? '点击收起' : '点击展开' }}</span>
          </div>
        </div>

        <!-- 分组内容区域 -->
        <div v-if="isGroupExpanded(type)" class="group-content">
          <div 
            v-for="rt in items" 
            :key="rt.id" 
            class="runtime-card"
          >
            <!-- 卡片头部 -->
            <div class="rt-card-top">
              <div class="rt-title-block">
                <span class="rt-title">{{ getRuntimeTypeName(type) }} {{ rt.version }}</span>
                <span v-if="rt.variant" class="rt-badge font-mono">{{ rt.variant.toUpperCase() }}</span>
                <span class="rt-badge font-mono">{{ rt.architecture }}</span>
                <p-tag 
                  v-if="rt.isDefault" 
                  severity="contrast" 
                  value="默认 CLI" 
                  icon="pi pi-star-fill" 
                  class="default-tag"
                />
              </div>

              <p-tag 
                :severity="getStatusSeverity(rt.status)" 
                :value="getStatusLabel(rt.status)"
                :icon="getStatusIcon(rt.status)"
              />
            </div>

            <!-- 卡片路径与引用信息 -->
            <div class="rt-info-row">
              <div class="rt-path font-mono" :title="rt.installPath">
                <i class="pi pi-folder"></i>
                <span>{{ rt.installPath }}</span>
              </div>
              <div class="rt-refs font-mono">
                <i class="pi pi-link"></i>
                <span>引用: {{ rt.refProjectsCount || 0 }} 个项目 · {{ rt.refServicesCount || 0 }} 个服务</span>
              </div>
            </div>

            <!-- 动态快捷操作按键 -->
            <div class="rt-actions">
              <!-- PHP 特有操作 -->
              <template v-if="type === 'php'">
                <p-button 
                  label="配置扩展" 
                  icon="pi pi-sliders-h" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="openExtensionsDialog(rt)"
                />
                <p-button 
                  v-if="!rt.isDefault"
                  label="设为默认" 
                  icon="pi pi-star" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  :loading="operatingId === rt.id"
                  @click="handleSetDefault(rt)"
                />
                <p-button 
                  label="创建服务" 
                  icon="pi pi-plus" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="handleCreateService(rt)"
                />
              </template>

              <!-- Nginx 特有操作 -->
              <template v-else-if="type === 'nginx'">
                <p-button 
                  label="测试配置" 
                  icon="pi pi-check-circle" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="handleHealthCheck(rt)"
                />
                <p-button 
                  label="创建服务" 
                  icon="pi pi-plus" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="handleCreateService(rt)"
                />
              </template>

              <!-- MySQL 特有操作 -->
              <template v-else-if="type === 'mysql'">
                <p-button 
                  label="初始化实例" 
                  icon="pi pi-cog" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="handleInitInstance(rt)"
                />
              </template>

              <!-- Redis 特有操作 -->
              <template v-else-if="type === 'redis'">
                <p-button 
                  label="测试连接" 
                  icon="pi pi-wifi" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="handleHealthCheck(rt)"
                />
                <p-button 
                  label="创建服务" 
                  icon="pi pi-plus" 
                  size="small" 
                  severity="secondary" 
                  outlined 
                  @click="handleCreateService(rt)"
                />
              </template>

              <!-- 通用健康检查 -->
              <p-button 
                v-if="type !== 'php' && type !== 'nginx' && type !== 'redis'"
                label="健康检查" 
                icon="pi pi-check-circle" 
                size="small" 
                severity="secondary" 
                outlined 
                @click="handleHealthCheck(rt)"
              />

              <!-- 通用卸载 -->
              <p-button 
                label="卸载" 
                icon="pi pi-trash" 
                size="small" 
                severity="danger" 
                outlined 
                class="btn-uninstall"
                @click="openUninstallDialog(rt)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 底部提示栏 -->
      <div class="bottom-tips-card">
        <div class="tip-line">
          <i class="pi pi-info-circle text-blue"></i>
          <span>所有运行环境均统一托管在本地 <code>runtimes/</code> 目录下，支持多版本并存与热拔插。</span>
        </div>
        <div class="tip-line">
          <i class="pi pi-shopping-bag text-indigo"></i>
          <span>需要安装其他环境或更多版本？随时前往
            <router-link to="/store" class="tip-link">【软件商店】</router-link> 在线下载部署。
          </span>
        </div>
      </div>
    </div>

    <!-- PHP 扩展配置弹窗 -->
    <h-modal 
      v-model="extDialogVisible" 
      :title="`PHP ${activeRuntime?.version} 扩展管理`"
      width="560px"
    >
      <div class="ext-dialog-body">
        <div class="ext-header-bar">
          <span class="ext-count">共发现 {{ extensions.length }} 个可用扩展</span>
          <p-tag severity="info" value="修改 php.ini 后需重启对应服务生效" />
        </div>

        <div v-if="loadingExt" class="loading-state">
          <i class="pi pi-spin pi-spinner"></i>
          <span>正在加载扩展列表...</span>
        </div>

        <div v-else class="ext-grid">
          <div 
            v-for="ext in extensions" 
            :key="ext.name" 
            class="ext-card"
            :class="{ enabled: ext.enabled }"
          >
            <div class="ext-info">
              <span class="ext-name font-mono">{{ ext.name }}</span>
              <span class="ext-status font-mono">{{ ext.enabled ? '已启用' : '未开启' }}</span>
            </div>
            <p-tag 
              :severity="ext.enabled ? 'success' : 'secondary'"
              :value="ext.enabled ? 'Active' : 'Off'"
            />
          </div>
        </div>
      </div>
      <template #footer>
        <p-button label="关闭" severity="secondary" outlined @click="extDialogVisible = false" />
      </template>
    </h-modal>

    <!-- 卸载确认对话框 -->
    <h-modal 
      v-model="uninstallDialogVisible" 
      title="卸载运行环境确认" 
      width="460px"
    >
      <div v-if="activeRuntime" class="uninstall-confirm-body">
        <div class="warning-callout">
          <i class="pi pi-exclamation-triangle warning-icon"></i>
          <div>
            <div class="warning-title">即将卸载环境：{{ getRuntimeTypeName(activeRuntime.runtimeType) }} {{ activeRuntime.version }}</div>
            <div class="warning-desc">此操作将从 DevEnv 中移除该运行时的注册信息。</div>
          </div>
        </div>

        <div v-if="(activeRuntime.refProjectsCount || 0) > 0 || (activeRuntime.refServicesCount || 0) > 0" class="conflict-alert">
          <i class="pi pi-times-circle"></i>
          <span>检测到有 {{ activeRuntime.refProjectsCount || 0 }} 个项目或 {{ activeRuntime.refServicesCount || 0 }} 个服务正在关联该环境，请先解除绑定！</span>
        </div>

        <label class="delete-files-checkbox">
          <input type="checkbox" v-model="deleteFilesCheckbox" />
          <span>同时永久删除物理文件目录（<code>{{ activeRuntime.installPath }}</code>）</span>
        </label>
      </div>

      <template #footer>
        <p-button label="取消" severity="secondary" outlined @click="uninstallDialogVisible = false" />
        <p-button 
          label="确认卸载" 
          severity="danger" 
          icon="pi pi-trash" 
          :loading="uninstalling" 
          :disabled="(activeRuntime?.refProjectsCount || 0) > 0 || (activeRuntime?.refServicesCount || 0) > 0"
          @click="confirmUninstall" 
        />
      </template>
    </h-modal>
  </h-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import HPage from '@/components/h-page/h-page.vue'
import HModal from '@/components/h-modal/h-modal.vue'
import { 
  listAllRuntimes, 
  uninstallRuntime, 
  setDefaultVersion, 
  healthCheck,
  getExtensions 
} from '@/api/runtimes'
import type { RuntimeVersion, PHPExtension } from '@/types'

defineOptions({
  name: 'RuntimesView'
})

const router = useRouter()
const loading = ref(false)
const operatingId = ref<number | null>(null)
const allRuntimes = ref<RuntimeVersion[]>([])
const expandedGroups = ref<string[]>(['php', 'nginx', 'mysql', 'redis', 'postgresql'])

// 扩展管理弹窗
const extDialogVisible = ref(false)
const activeRuntime = ref<RuntimeVersion | null>(null)
const extensions = ref<PHPExtension[]>([])
const loadingExt = ref(false)

// 卸载弹窗
const uninstallDialogVisible = ref(false)
const deleteFilesCheckbox = ref(false)
const uninstalling = ref(false)

const groupedRuntimes = computed(() => {
  const map: Record<string, RuntimeVersion[]> = {}
  allRuntimes.value.forEach(rt => {
    const t = rt.runtimeType.toLowerCase()
    if (!map[t]) {
      map[t] = []
    }
    map[t].push(rt)
  })
  return map
})

function getRuntimeTypeName(type: string): string {
  const names: Record<string, string> = {
    php: 'PHP',
    nginx: 'Nginx',
    mysql: 'MySQL',
    postgresql: 'PostgreSQL',
    redis: 'Redis',
    apache: 'Apache',
    nodejs: 'Node.js',
  }
  return names[type.toLowerCase()] || type.toUpperCase()
}

function getRuntimeBrandColor(type: string): string {
  const colors: Record<string, string> = {
    php: '#777BB4',
    nginx: '#009639',
    mysql: '#4479A1',
    redis: '#DC382D',
    postgresql: '#336791',
    apache: '#D22128',
    nodejs: '#339933',
  }
  return colors[type.toLowerCase()] || '#64748b'
}

function getRuntimeIcon(type: string): string {
  const icons: Record<string, string> = {
    php: 'pi pi-code',
    nginx: 'pi pi-server',
    mysql: 'pi pi-database',
    redis: 'pi pi-bolt',
    postgresql: 'pi pi-database',
    apache: 'pi pi-compass',
    nodejs: 'pi pi-code',
  }
  return icons[type.toLowerCase()] || 'pi pi-box'
}

function isGroupExpanded(type: string): boolean {
  return expandedGroups.value.includes(type)
}

function toggleGroup(type: string) {
  const idx = expandedGroups.value.indexOf(type)
  if (idx > -1) {
    expandedGroups.value.splice(idx, 1)
  } else {
    expandedGroups.value.push(type)
  }
}

function getStatusSeverity(status: string) {
  switch (status) {
    case 'available':
    case 'installed':
      return 'success'
    case 'uninitialized':
      return 'warn'
    case 'in_use':
      return 'info'
    case 'error':
    case 'failed':
      return 'danger'
    default:
      return 'secondary'
  }
}

function getStatusLabel(status: string) {
  switch (status) {
    case 'available':
    case 'installed':
      return '可用'
    case 'uninitialized':
      return '未初始化'
    case 'in_use':
      return '使用中'
    case 'installing':
      return '安装中'
    case 'error':
    case 'failed':
      return '不可用'
    default:
      return '就绪'
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'available':
    case 'installed':
      return 'pi pi-check'
    case 'uninitialized':
      return 'pi pi-exclamation-triangle'
    case 'error':
    case 'failed':
      return 'pi pi-times'
    default:
      return 'pi pi-info-circle'
  }
}

async function loadRuntimes() {
  loading.value = true
  try {
    allRuntimes.value = await listAllRuntimes()
  } catch (err: any) {
    toast.error('加载运行环境失败: ' + (err?.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

async function handleSetDefault(rt: RuntimeVersion) {
  operatingId.value = rt.id
  try {
    await setDefaultVersion(rt.id)
    toast.success(`已将 PHP ${rt.version} 设为默认全局 CLI`)
    await loadRuntimes()
  } catch (err: any) {
    toast.error('设置默认版本失败: ' + (err?.message || '未知错误'))
  } finally {
    operatingId.value = null
  }
}

async function handleHealthCheck(rt: RuntimeVersion) {
  try {
    const res = await healthCheck(rt.id)
    if (res.healthy) {
      toast.success(res.message)
    } else {
      toast.warn(res.message)
    }
  } catch (err: any) {
    toast.error('健康检查失败: ' + (err?.message || '未知错误'))
  }
}

function handleCreateService(rt: RuntimeVersion) {
  router.push({
    path: '/services',
    query: { createForRuntime: String(rt.id), runtimeType: rt.runtimeType }
  })
}

function handleInitInstance(rt: RuntimeVersion) {
  toast.info(`正在准备初始化 ${rt.runtimeType.toUpperCase()} 实例数据目录...`)
  router.push('/services')
}

async function openExtensionsDialog(rt: RuntimeVersion) {
  activeRuntime.value = rt
  extDialogVisible.value = true
  loadingExt.value = true
  try {
    extensions.value = await getExtensions(rt.id)
  } catch (err: any) {
    toast.error('获取扩展列表失败: ' + (err?.message || '未知错误'))
  } finally {
    loadingExt.value = false
  }
}

function openUninstallDialog(rt: RuntimeVersion) {
  activeRuntime.value = rt
  deleteFilesCheckbox.value = false
  uninstallDialogVisible.value = true
}

async function confirmUninstall() {
  if (!activeRuntime.value) return
  uninstalling.value = true
  try {
    const res = await uninstallRuntime(activeRuntime.value.id, deleteFilesCheckbox.value)
    toast.success(res.message || '卸载成功')
    uninstallDialogVisible.value = false
    await loadRuntimes()
  } catch (err: any) {
    toast.error('卸载失败: ' + (err?.message || '未知错误'))
  } finally {
    uninstalling.value = false
  }
}

onMounted(() => {
  loadRuntimes()
})
</script>

<style scoped>
.groups-container {
  display: flex;
  flex-direction: column;
  gap: var(--h-padding);
}

.runtime-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  cursor: pointer;
  user-select: none;
  transition: none !important;
}

.group-header:hover {
  background: #f8fafc;
  border-color: var(--border-hover);
}

.group-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.group-icon-badge {
  width: 26px;
  height: 26px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 0.85rem;
}

.group-name {
  font-size: 0.95rem;
  color: var(--text-primary);
}

.count-tag {
  font-size: 0.725rem;
  padding: 1px 6px;
}

.group-header-right {
  display: flex;
  align-items: center;
}

.collapse-hint {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.group-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-left: 6px;
}

.runtime-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  padding: var(--h-padding);
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: none !important;
}

.runtime-card:hover {
  border-color: var(--border-hover);
}

.rt-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rt-title-block {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rt-title {
  font-size: 0.95rem;
  color: var(--text-primary);
}

.rt-badge {
  font-size: 0.725rem;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
}

.default-tag {
  font-size: 0.7rem;
  padding: 1px 6px;
}

.rt-info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.rt-path {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: var(--bg-app);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rt-refs {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.775rem;
  color: var(--text-muted);
  background: var(--bg-app);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  flex-shrink: 0;
}

.rt-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  border-top: 1px solid var(--border-color);
  padding-top: 10px;
}

.btn-uninstall {
  margin-left: auto;
}

/* 底部提示 */
.bottom-tips-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 14px;
  background: #f8fafc;
  border: 1px dashed var(--border-color);
  border-radius: var(--h-radius);
}

.tip-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.tip-link {
  color: var(--p-primary-color);
  text-decoration: none;
}
.tip-link:hover {
  text-decoration: underline;
}

.text-blue { color: #2563eb; }
.text-indigo { color: #4f46e5; }

/* 扩展弹窗 */
.ext-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ext-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ext-count {
  font-size: 0.825rem;
  color: var(--text-muted);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 32px;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.ext-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  max-height: 320px;
  overflow-y: auto;
}

.ext-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.ext-card.enabled {
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.ext-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ext-name {
  font-size: 0.85rem;
  color: var(--text-primary);
}

.ext-status {
  font-size: 0.725rem;
  color: var(--text-muted);
}

/* 卸载确认 */
.uninstall-confirm-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.warning-callout {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 12px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: var(--h-radius);
}

.warning-icon {
  font-size: 1.3rem;
  color: #d97706;
  flex-shrink: 0;
  margin-top: 2px;
}

.warning-title {
  font-size: 0.9rem;
  color: #92400e;
}

.warning-desc {
  font-size: 0.775rem;
  color: #b45309;
  margin-top: 2px;
}

.conflict-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #b91c1c;
  font-size: 0.8rem;
  border-radius: var(--radius-sm);
}

.delete-files-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.825rem;
  color: var(--text-primary);
  cursor: pointer;
  margin-top: 4px;
}

.empty-runtimes {
  text-align: center;
  padding: 60px 0;
  color: var(--text-dim);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 12px;
}

.empty-title {
  font-size: 1rem;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.empty-subtitle {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 16px;
}
</style>
