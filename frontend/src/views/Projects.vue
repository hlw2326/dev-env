<template>
  <div class="projects-page fade-in">
    <!-- Action Bar -->
    <div class="page-header glass-panel">
      <div class="header-left">
        <h2>项目管理</h2>
        <p>登记与关联本地 PHP / Web 项目目录及运行版本</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="showImportModal = true">
          <FolderInputIcon class="btn-icon" />
          导入已有项目
        </button>
        <button class="btn btn-primary" @click="showCreateModal = true">
          <PlusIcon class="btn-icon" />
          新建项目
        </button>
      </div>
    </div>

    <!-- Filter & Table -->
    <div class="table-container glass-panel">
      <div class="table-toolbar">
        <input 
          v-model="searchQuery" 
          type="text" 
          class="form-input search-input" 
          placeholder="搜索项目名称或路径..." 
        />
        <button class="btn btn-secondary btn-sm" @click="loadProjects">
          <RefreshCwIcon class="btn-icon" :class="{ 'spin-icon': loading }" />
          刷新
        </button>
      </div>

      <div v-if="filteredProjects.length === 0" class="empty-state">
        <span v-if="loading">加载项目中...</span>
        <span v-else>暂无匹配的项目，点击上方按钮新建或导入项目</span>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>项目名称</th>
            <th>物理路径</th>
            <th>关联 PHP</th>
            <th>状态</th>
            <th>创建时间</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="proj in filteredProjects" :key="proj.id">
            <td class="font-mono text-dim">#{{ proj.id }}</td>
            <td class="font-bold text-main">{{ proj.name }}</td>
            <td class="font-mono text-sm text-muted cell-path" :title="proj.path">{{ proj.path }}</td>
            <td>
              <span v-if="proj.phpVersionId" class="badge badge-primary">
                PHP {{ getPHPVersionLabel(proj.phpVersionId) }}
              </span>
              <span v-else class="badge badge-neutral">默认环境</span>
            </td>
            <td>
              <span class="badge" :class="proj.status === 'ready' ? 'badge-success' : 'badge-danger'">
                {{ proj.status === 'ready' ? '正常就绪' : '路径缺失' }}
              </span>
            </td>
            <td class="text-sm text-dim">{{ formatDate(proj.createdAt) }}</td>
            <td class="text-right">
              <button class="btn btn-danger btn-sm" @click="handleDelete(proj.id)">
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
        <h3>新建项目</h3>
        <p class="modal-subtitle">创建新项目登记记录并建立物理目录</p>

        <form @submit.prevent="submitCreate">
          <div class="form-group">
            <label class="form-label">项目名称</label>
            <input v-model="createForm.name" type="text" class="form-input" required placeholder="如 my-laravel-app" />
          </div>

          <div class="form-group">
            <label class="form-label">物理目录路径</label>
            <input 
              v-model="createForm.path" 
              type="text" 
              class="form-input" 
              required 
              placeholder="如 G:\www\my-laravel-app"
              @input="onPathChange"
            />
            <div v-if="pathFeedback" class="path-feedback" :class="pathFeedback.valid ? 'text-success' : 'text-danger'">
              {{ pathFeedback.valid ? '✓ 路径有效' : pathFeedback.errors.join('; ') }}
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">绑定 PHP 版本</label>
            <select v-model="createForm.phpVersionId" class="form-input">
              <option :value="null">默认运行版本</option>
              <option v-for="php in phpVersions" :key="php.id" :value="php.id">
                PHP {{ php.version }} ({{ php.architecture }})
              </option>
            </select>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">创建</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Import Modal -->
    <div v-if="showImportModal" class="modal-overlay" @click.self="showImportModal = false">
      <div class="modal-content">
        <h3>导入已有项目</h3>
        <p class="modal-subtitle">将本地已有代码目录快速纳入 DevEnv 管理</p>

        <form @submit.prevent="submitImport">
          <div class="form-group">
            <label class="form-label">已有目录绝对路径</label>
            <input v-model="importPath" type="text" class="form-input" required placeholder="如 D:\projects\shop" />
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showImportModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">确认导入</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus as PlusIcon, FolderInput as FolderInputIcon, Trash2 as Trash2Icon, RefreshCw as RefreshCwIcon } from 'lucide-vue-next'
import { listProjects, createProject, importProject, deleteProject, validateProjectPath } from '../api/projects'
import { listVersions } from '../api/php'
import type { Project, PHPVersion, ProjectPathValidation } from '../types'

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
    showCreateModal.value = false
    createForm.value = { name: '', path: '', phpVersionId: null }
    pathFeedback.value = null
    await loadProjects()
  } catch (err: any) {
    alert(err?.message || '创建项目失败')
  } finally {
    submitting.value = false
  }
}

async function submitImport() {
  submitting.value = true
  try {
    await importProject(importPath.value)
    showImportModal.value = false
    importPath.value = ''
    await loadProjects()
  } catch (err: any) {
    alert(err?.message || '导入项目失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除该项目的管理记录吗？物理项目目录将保留。')) {
    return
  }
  try {
    await deleteProject(id)
    await loadProjects()
  } catch (err: any) {
    alert(err?.message || '删除失败')
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
  gap: 12px;
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
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-right {
  text-align: right;
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
.text-success {
  color: #34d399;
}
.text-danger {
  color: #fb7185;
}

.modal-subtitle {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 18px;
}

.form-group {
  margin-bottom: 14px;
}

.path-feedback {
  font-size: 0.75rem;
  margin-top: 4px;
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
