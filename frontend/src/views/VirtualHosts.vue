<template>
  <div class="hosts-page fade-in">
    <!-- Header -->
    <div class="page-header glass-panel">
      <div class="header-left">
        <h2>虚拟主机 / 域名管理</h2>
        <p>配置本地域名映射、绑定 Web 网站根目录与端口分发</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="showCreateModal = true">
          <PlusIcon class="btn-icon" />
          添加虚拟主机
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-container glass-panel">
      <div class="table-toolbar">
        <input 
          v-model="searchQuery" 
          type="text" 
          class="form-input search-input" 
          placeholder="搜索本地域名或项目..." 
        />
        <button class="btn btn-secondary btn-sm" @click="loadHosts">
          <RefreshCwIcon class="btn-icon" :class="{ 'spin-icon': loading }" />
          刷新
        </button>
      </div>

      <div v-if="filteredHosts.length === 0" class="empty-state">
        <span v-if="loading">加载虚拟主机中...</span>
        <span v-else>暂无虚拟主机，点击上方按钮添加本地域名绑定</span>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>本地域名</th>
            <th>关联项目</th>
            <th>网站根目录 (Document Root)</th>
            <th>服务器</th>
            <th>端口</th>
            <th>状态</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="host in filteredHosts" :key="host.id">
            <td class="font-mono text-dim">#{{ host.id }}</td>
            <td class="font-bold font-mono text-main">{{ host.domain }}</td>
            <td>
              <span class="badge badge-primary">{{ getProjectName(host.projectId) }}</span>
            </td>
            <td class="font-mono text-sm text-muted cell-path" :title="host.documentRoot">
              {{ host.documentRoot }}
            </td>
            <td>
              <span class="badge badge-neutral font-mono">{{ host.webServerType.toUpperCase() }}</span>
            </td>
            <td class="font-mono text-dim">{{ host.port }}</td>
            <td>
              <span class="badge" :class="host.enabled ? 'badge-success' : 'badge-neutral'">
                {{ host.enabled ? '已启用' : '已禁用' }}
              </span>
            </td>
            <td class="text-right">
              <button 
                class="btn btn-secondary btn-sm mr-2"
                @click="toggleHostStatus(host)"
              >
                {{ host.enabled ? '禁用' : '启用' }}
              </button>
              <button class="btn btn-danger btn-sm" @click="handleDelete(host.id)">
                <Trash2Icon class="btn-icon" />
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h3>添加虚拟主机</h3>
        <p class="modal-subtitle">绑定自定义本地域名到指定的本地项目目录</p>

        <form @submit.prevent="submitCreate">
          <div class="form-group">
            <label class="form-label">关联项目</label>
            <select v-model="createForm.projectId" class="form-input" required @change="onProjectSelect">
              <option disabled :value="0">请选择项目</option>
              <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }} ({{ p.path }})</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">本地域名</label>
            <input 
              v-model="createForm.domain" 
              type="text" 
              class="form-input font-mono" 
              required 
              placeholder="如 app.test 或 my-site.local"
            />
          </div>

          <div class="form-group">
            <label class="form-label">网站根目录 (Document Root)</label>
            <input 
              v-model="createForm.documentRoot" 
              type="text" 
              class="form-input font-mono" 
              required 
              placeholder="如 G:\www\my-laravel-app\public"
            />
          </div>

          <div class="form-row">
            <div class="form-group flex-1">
              <label class="form-label">Web 服务器</label>
              <select v-model="createForm.webServerType" class="form-input">
                <option value="nginx">Nginx</option>
                <option value="apache">Apache</option>
              </select>
            </div>
            <div class="form-group flex-1">
              <label class="form-label">监听端口</label>
              <input v-model.number="createForm.port" type="number" class="form-input" required min="1" max="65535" />
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">创建绑定</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus as PlusIcon, RefreshCw as RefreshCwIcon, Trash2 as Trash2Icon } from 'lucide-vue-next'
import { listVirtualHosts, createVirtualHost, deleteVirtualHost, enableVirtualHost, disableVirtualHost } from '../api/hosts'
import { listProjects } from '../api/projects'
import type { VirtualHost, Project } from '../types'

const hosts = ref<VirtualHost[]>([])
const projects = ref<Project[]>([])
const searchQuery = ref('')
const loading = ref(false)
const submitting = ref(false)

const showCreateModal = ref(false)
const createForm = ref<{
  projectId: number
  domain: string
  documentRoot: string
  webServerType: 'nginx' | 'apache'
  port: number
}>({
  projectId: 0,
  domain: '',
  documentRoot: '',
  webServerType: 'nginx',
  port: 80,
})

const filteredHosts = computed(() => {
  if (!searchQuery.value) return hosts.value
  const q = searchQuery.value.toLowerCase()
  return hosts.value.filter(h => h.domain.toLowerCase().includes(q) || h.documentRoot.toLowerCase().includes(q))
})

function getProjectName(projectId: number) {
  const found = projects.value.find(p => p.id === projectId)
  return found ? found.name : `项目 #${projectId}`
}

function onProjectSelect() {
  const p = projects.value.find(p => p.id === createForm.value.projectId)
  if (p) {
    if (!createForm.value.domain) {
      createForm.value.domain = `${p.name.toLowerCase().replace(/[^a-z0-9]/g, '')}.local`
    }
    if (!createForm.value.documentRoot) {
      createForm.value.documentRoot = p.path
    }
  }
}

async function loadHosts() {
  loading.value = true
  try {
    const [hList, pList] = await Promise.all([
      listVirtualHosts(),
      listProjects(),
    ])
    hosts.value = hList
    projects.value = pList
  } finally {
    loading.value = false
  }
}

async function submitCreate() {
  if (!createForm.value.projectId) {
    alert('请选择关联项目')
    return
  }
  submitting.value = true
  try {
    await createVirtualHost({
      projectId: createForm.value.projectId,
      domain: createForm.value.domain,
      documentRoot: createForm.value.documentRoot,
      webServerType: createForm.value.webServerType,
      port: createForm.value.port,
    })
    showCreateModal.value = false
    createForm.value = {
      projectId: 0,
      domain: '',
      documentRoot: '',
      webServerType: 'nginx',
      port: 80,
    }
    await loadHosts()
  } catch (err: any) {
    alert(err?.message || '创建虚拟主机失败')
  } finally {
    submitting.value = false
  }
}

async function toggleHostStatus(host: VirtualHost) {
  try {
    if (host.enabled) {
      await disableVirtualHost(host.id)
    } else {
      await enableVirtualHost(host.id)
    }
    await loadHosts()
  } catch (err: any) {
    alert(err?.message || '切换状态失败')
  }
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除该虚拟主机配置吗？')) {
    return
  }
  try {
    await deleteVirtualHost(id)
    await loadHosts()
  } catch (err: any) {
    alert(err?.message || '删除失败')
  }
}

onMounted(() => {
  loadHosts()
})
</script>

<style scoped>
.hosts-page {
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

.table-container {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-input {
  max-width: 320px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.data-table th {
  padding: 10px 12px;
  font-size: 0.775rem;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border-subtle);
}

.data-table td {
  padding: 12px;
  font-size: 0.85rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.data-table tr:hover td {
  background: rgba(255, 255, 255, 0.02);
}

.cell-path {
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-right {
  text-align: right;
}

.mr-2 {
  margin-right: 8px;
}

.font-mono {
  font-family: var(--font-mono);
}
.font-bold {
  font-weight: 600;
}
.text-main {
  color: var(--text-main);
}
.text-muted {
  color: var(--text-muted);
}
.text-dim {
  color: var(--text-dim);
}

.modal-subtitle {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 18px;
}

.form-group {
  margin-bottom: 14px;
}

.form-row {
  display: flex;
  gap: 12px;
}

.flex-1 {
  flex: 1;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.empty-state {
  text-align: center;
  color: var(--text-dim);
  padding: 40px;
  font-size: 0.875rem;
}
</style>
