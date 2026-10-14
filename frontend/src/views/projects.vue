<template>
  <h-page title="项目管理" desc="登记本地 PHP / Web 项目目录并关联指定运行版本">
    <template #extra>
      <p-button 
        label="导入已有项目" 
        icon="pi pi-folder-open" 
        severity="secondary" 
        outlined 
        @click="showImportModal = true" 
      />
      <p-button 
        label="新建项目" 
        icon="pi pi-plus" 
        @click="showCreateModal = true" 
      />
    </template>

    <!-- Table Card -->
    <h-card noPadding>
      <div class="table-header-tools">
        <span class="p-input-icon-left">
          <p-input-text v-model="searchQuery" placeholder="搜索项目名称或路径..." class="search-input" />
        </span>
        <p-button 
          icon="pi pi-refresh" 
          label="刷新" 
          severity="secondary" 
          text 
          :loading="loading" 
          @click="loadProjects" 
        />
      </div>

      <p-data-table 
        :value="filteredProjects" 
        stripedRows 
        responsiveLayout="scroll"
        :paginator="filteredProjects.length > 10" 
        :rows="10"
        class="p-datatable-sm"
      >
        <template #empty>
          <div class="empty-state">
            <span v-if="loading">加载项目中...</span>
            <span v-else>暂无项目，点击右上角新建或导入已有项目</span>
          </div>
        </template>

        <p-column field="id" header="ID" style="width: 70px">
          <template #body="{ data }">
            <span class="font-mono text-dim">#{{ data.id }}</span>
          </template>
        </p-column>

        <p-column field="name" header="项目名称" style="width: 200px">
          <template #body="{ data }">
            <span class="project-title">{{ data.name }}</span>
          </template>
        </p-column>

        <p-column field="path" header="物理路径">
          <template #body="{ data }">
            <span class="font-mono text-sm text-path" :title="data.path">{{ data.path }}</span>
          </template>
        </p-column>

        <p-column field="phpVersionId" header="关联 PHP" style="width: 140px">
          <template #body="{ data }">
            <p-tag 
              v-if="data.phpVersionId" 
              severity="info" 
              :value="'PHP ' + getPHPVersionLabel(data.phpVersionId)" 
            />
            <p-tag v-else severity="secondary" value="默认环境" />
          </template>
        </p-column>

        <p-column field="status" header="状态" style="width: 110px">
          <template #body="{ data }">
            <p-tag 
              :severity="data.status === 'ready' ? 'success' : 'danger'"
              :value="data.status === 'ready' ? '就绪' : '路径缺失'" 
            />
          </template>
        </p-column>

        <p-column field="createdAt" header="创建时间" style="width: 140px">
          <template #body="{ data }">
            <span class="text-sm text-muted">{{ formatDate(data.createdAt) }}</span>
          </template>
        </p-column>

        <p-column header="操作" style="width: 100px; text-align: right">
          <template #body="{ data }">
            <p-button 
              icon="pi pi-trash" 
              severity="danger" 
              text 
              rounded 
              size="small" 
              @click="handleDelete(data.id)" 
            />
          </template>
        </p-column>
      </p-data-table>
    </h-card>

    <!-- Create Modal -->
    <h-modal 
      v-model="showCreateModal" 
      title="新建项目" 
      width="480px"
      ok-text="确认创建"
      :loading="submitting"
      @ok="submitCreate"
    >
      <form @submit.prevent="submitCreate">
        <div class="form-field">
          <label>项目名称</label>
          <p-input-text v-model="createForm.name" required placeholder="如 my-laravel-app" />
        </div>

        <div class="form-field">
          <label>物理目录路径</label>
          <p-input-text 
            v-model="createForm.path" 
            required 
            placeholder="如 G:\www\my-laravel-app" 
            @input="onPathChange"
          />
          <div v-if="pathFeedback" class="form-feedback" :class="pathFeedback.valid ? 'valid' : 'invalid'">
            {{ pathFeedback.valid ? '✓ 路径有效' : pathFeedback.errors.join('; ') }}
          </div>
        </div>

        <div class="form-field">
          <label>绑定 PHP 版本</label>
          <p-select 
            v-model="createForm.phpVersionId" 
            :options="phpOptions" 
            optionLabel="label" 
            optionValue="value" 
            placeholder="选择绑定的 PHP 版本" 
          />
        </div>
      </form>
    </h-modal>

    <!-- Import Modal -->
    <h-modal 
      v-model="showImportModal" 
      title="导入已有项目" 
      width="480px"
      ok-text="确认导入"
      :loading="submitting"
      @ok="submitImport"
    >
      <form @submit.prevent="submitImport">
        <div class="form-field">
          <label>已有目录物理绝对路径</label>
          <p-input-text v-model="importPath" required placeholder="如 D:\projects\shop" />
        </div>
      </form>
    </h-modal>
  </h-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import HPage from '@/components/h-page/h-page.vue'
import HCard from '@/components/h-card/h-card.vue'
import HModal from '@/components/h-modal/h-modal.vue'
import { listProjects, createProject, importProject, deleteProject, validateProjectPath } from '@/api/projects'
import { listVersions } from '@/api/php'
import type { Project, PHPVersion, ProjectPathValidation } from '@/types'

defineOptions({
  name: 'Projects'
})

const projects = ref<Project[]>([])
const phpVersions = ref<PHPVersion[]>([])
const searchQuery = ref('')
const loading = ref(false)
const submitting = ref(false)

const showCreateModal = ref(false)
const showImportModal = ref(false)
const importPath = ref('')

const createForm = ref<{
  name: string
  path: string
  phpVersionId: number | null
}>({
  name: '',
  path: '',
  phpVersionId: null,
})

const pathFeedback = ref<ProjectPathValidation | null>(null)

const phpOptions = computed(() => {
  const options = [{ label: '默认环境 (不指定)', value: null as number | null }]
  for (const v of phpVersions.value) {
    options.push({
      label: `PHP ${v.version} (${v.architecture})`,
      value: v.id,
    })
  }
  return options
})

const filteredProjects = computed(() => {
  if (!searchQuery.value) return projects.value
  const q = searchQuery.value.toLowerCase()
  return projects.value.filter(p => p.name.toLowerCase().includes(q) || p.path.toLowerCase().includes(q))
})

function getPHPVersionLabel(phpVersionId: number) {
  const found = phpVersions.value.find(v => v.id === phpVersionId)
  return found ? found.version : String(phpVersionId)
}

function formatDate(iso: string) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString()
}

async function loadProjects() {
  loading.value = true
  try {
    const [pList, vList] = await Promise.all([
      listProjects(),
      listVersions(),
    ])
    projects.value = pList
    phpVersions.value = vList
  } finally {
    loading.value = false
  }
}

async function onPathChange() {
  if (!createForm.value.path) {
    pathFeedback.value = null
    return
  }
  pathFeedback.value = await validateProjectPath(createForm.value.path)
}

async function submitCreate() {
  submitting.value = true
  try {
    await createProject({
      name: createForm.value.name,
      path: createForm.value.path,
      phpVersionId: createForm.value.phpVersionId,
    })
    toast.success('项目已成功创建')
    showCreateModal.value = false
    createForm.value = { name: '', path: '', phpVersionId: null }
    pathFeedback.value = null
    await loadProjects()
  } catch (err: any) {
    toast.error(err?.message || '创建项目失败')
  } finally {
    submitting.value = false
  }
}

async function submitImport() {
  submitting.value = true
  try {
    await importProject(importPath.value)
    toast.success('项目已成功导入')
    showImportModal.value = false
    importPath.value = ''
    await loadProjects()
  } catch (err: any) {
    toast.error(err?.message || '导入项目失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除该项目的登记记录吗？项目物理目录将完整保留。')) {
    return
  }
  try {
    await deleteProject(id)
    toast.info('项目管理记录已删除')
    await loadProjects()
  } catch (err: any) {
    toast.error(err?.message || '删除失败')
  }
}

onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
.projects-page {
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

.project-title {
  font-weight: 400;
  color: var(--text-main);
}

.text-path {
  max-width: 380px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.text-dim { color: #94a3b8; }
.text-muted { color: #64748b; }
.text-sm { font-size: 0.8rem; }

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
