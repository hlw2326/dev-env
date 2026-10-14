<template>
  <div class="hosts-page">
    <!-- Header -->
    <div class="page-header-bar">
      <div>
        <h1 class="page-title">虚拟主机 / 域名管理</h1>
        <p class="page-subtitle">配置本地域名映射、绑定 Web 网站根目录与端口分发</p>
      </div>
      <div class="header-actions">
        <Button 
          label="添加虚拟主机" 
          icon="pi pi-plus" 
          @click="showCreateModal = true" 
        />
      </div>
    </div>

    <!-- Table Card -->
    <div class="surface-card table-card">
      <div class="table-header-tools">
        <InputText v-model="searchQuery" placeholder="搜索本地域名或项目..." class="search-input" />
        <Button 
          icon="pi pi-refresh" 
          label="刷新" 
          severity="secondary" 
          text 
          :loading="loading" 
          @click="loadHosts" 
        />
      </div>

      <DataTable 
        :value="filteredHosts" 
        stripedRows 
        responsiveLayout="scroll"
        :paginator="filteredHosts.length > 10" 
        :rows="10"
        class="p-datatable-sm"
      >
        <template #empty>
          <div class="empty-state">
            <span v-if="loading">加载虚拟主机中...</span>
            <span v-else>暂无虚拟主机，点击右上角添加本地域名映射</span>
          </div>
        </template>

        <Column field="id" header="ID" style="width: 70px">
          <template #body="{ data }">
            <span class="font-mono text-dim">#{{ data.id }}</span>
          </template>
        </Column>

        <Column field="domain" header="本地域名" style="width: 180px">
          <template #body="{ data }">
            <span class="font-bold font-mono text-blue">{{ data.domain }}</span>
          </template>
        </Column>

        <Column field="projectId" header="关联项目" style="width: 160px">
          <template #body="{ data }">
            <Tag severity="info" :value="getProjectName(data.projectId)" />
          </template>
        </Column>

        <Column field="documentRoot" header="网站根目录 (Document Root)">
          <template #body="{ data }">
            <span class="font-mono text-sm text-path" :title="data.documentRoot">{{ data.documentRoot }}</span>
          </template>
        </Column>

        <Column field="webServerType" header="服务器" style="width: 110px">
          <template #body="{ data }">
            <Tag severity="secondary" :value="data.webServerType.toUpperCase()" />
          </template>
        </Column>

        <Column field="port" header="端口" style="width: 90px">
          <template #body="{ data }">
            <span class="font-mono text-muted">{{ data.port }}</span>
          </template>
        </Column>

        <Column field="enabled" header="状态" style="width: 100px">
          <template #body="{ data }">
            <Tag 
              :severity="data.enabled ? 'success' : 'secondary'"
              :value="data.enabled ? '已启用' : '已禁用'" 
            />
          </template>
        </Column>

        <Column header="操作" style="width: 140px; text-align: right">
          <template #body="{ data }">
            <Button 
              :label="data.enabled ? '禁用' : '启用'" 
              severity="secondary" 
              text 
              size="small"
              @click="toggleHostStatus(data)" 
            />
            <Button 
              icon="pi pi-trash" 
              severity="danger" 
              text 
              rounded 
              size="small"
              @click="handleDelete(data.id)" 
            />
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Create Dialog -->
    <Dialog 
      v-model:visible="showCreateModal" 
      modal 
      header="添加虚拟主机" 
      :style="{ width: '480px' }"
    >
      <form @submit.prevent="submitCreate">
        <div class="form-field">
          <label>关联项目</label>
          <Select 
            v-model="createForm.projectId" 
            :options="projectOptions" 
            optionLabel="label" 
            optionValue="value" 
            placeholder="请选择项目"
            required
            @change="onProjectSelect"
          />
        </div>

        <div class="form-field">
          <label>本地域名</label>
          <InputText 
            v-model="createForm.domain" 
            class="font-mono" 
            required 
            placeholder="如 app.test 或 my-site.local" 
          />
        </div>

        <div class="form-field">
          <label>网站根目录 (Document Root)</label>
          <InputText 
            v-model="createForm.documentRoot" 
            class="font-mono" 
            required 
            placeholder="如 G:\www\my-laravel-app\public" 
          />
        </div>

        <div class="form-row">
          <div class="form-field flex-1">
            <label>Web 服务器</label>
            <Select 
              v-model="createForm.webServerType" 
              :options="serverOptions" 
              optionLabel="label" 
              optionValue="value" 
            />
          </div>
          <div class="form-field flex-1">
            <label>监听端口</label>
            <input 
              v-model.number="createForm.port" 
              type="number" 
              class="p-inputtext p-component font-mono" 
              required 
              min="1" 
              max="65535" 
            />
          </div>
        </div>

        <div class="dialog-footer">
          <Button label="取消" severity="secondary" text @click="showCreateModal = false" />
          <Button label="确认创建" type="submit" :loading="submitting" />
        </div>
      </form>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { listVirtualHosts, createVirtualHost, deleteVirtualHost, enableVirtualHost, disableVirtualHost } from '../api/hosts'
import { listProjects } from '../api/projects'
import type { VirtualHost, Project } from '../types'

const toast = useToast()
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

const serverOptions = [
  { label: 'Nginx', value: 'nginx' },
  { label: 'Apache', value: 'apache' },
]

const projectOptions = computed(() => {
  return projects.value.map(p => ({
    label: `${p.name} (${p.path})`,
    value: p.id,
  }))
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
    toast.add({ severity: 'warn', summary: '提示', detail: '请选择关联的项目', life: 2500 })
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
    toast.add({ severity: 'success', summary: '成功', detail: '虚拟主机已创建', life: 2500 })
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
    toast.add({ severity: 'error', summary: '错误', detail: err?.message || '创建失败', life: 3500 })
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
    toast.add({ severity: 'error', summary: '错误', detail: err?.message || '切换失败', life: 3500 })
  }
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除该虚拟主机配置吗？')) return
  try {
    await deleteVirtualHost(id)
    toast.add({ severity: 'info', summary: '已删除', detail: '虚拟主机已删除', life: 2500 })
    await loadHosts()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: '错误', detail: err?.message || '删除失败', life: 3500 })
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
}

.table-card {
  padding: 16px;
}

.table-header-tools {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.search-input {
  width: 280px;
}

.text-blue { color: #2563eb; }
.text-dim { color: #94a3b8; }
.text-muted { color: #64748b; }
.text-sm { font-size: 0.8rem; }

.text-path {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.form-row {
  display: flex;
  gap: 12px;
}

.flex-1 {
  flex: 1;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
