<template>
  <div class="tasks-page fade-in">
    <!-- Header -->
    <div class="page-header glass-panel">
      <div class="header-left">
        <h2>任务与后台作业中心</h2>
        <p>查看环境安装、备份、恢复及异步处理任务的进度与状态</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="loadTasks">
          <RefreshCwIcon class="btn-icon" :class="{ 'spin-icon': loading }" />
          刷新
        </button>
      </div>
    </div>

    <!-- Tasks List -->
    <div class="table-container glass-panel">
      <div v-if="tasks.length === 0" class="empty-state">
        <span v-if="loading">加载任务中...</span>
        <span v-else>当前没有正在运行或历史记录的任务</span>
      </div>

      <div v-else class="tasks-list">
        <div v-for="t in tasks" :key="t.id" class="task-card">
          <div class="task-header-row">
            <div class="task-title-group">
              <span class="task-type-badge font-mono">{{ t.taskType }}</span>
              <span class="task-id font-mono">#{{ t.id }}</span>
              <span class="task-msg">{{ t.message || '处理中...' }}</span>
            </div>

            <div class="task-status-group">
              <span class="badge" :class="getStatusBadgeClass(t.status)">
                {{ getStatusLabel(t.status) }}
              </span>

              <button 
                v-if="t.status === 'running' || t.status === 'queued'" 
                class="btn btn-danger btn-sm" 
                @click="handleCancel(t.id)"
              >
                取消任务
              </button>
            </div>
          </div>

          <!-- Progress Bar -->
          <div class="progress-bar-wrap">
            <div class="progress-bar-fill" :style="{ width: `${t.progress}%` }"></div>
          </div>

          <div class="task-footer-row">
            <span class="text-dim text-xs">创建于: {{ formatDate(t.createdAt) }}</span>
            <span class="text-dim text-xs font-mono">{{ t.progress }}%</span>
          </div>

          <div v-if="t.errorMessage" class="task-error-box font-mono">
            错误: [{{ t.errorCode }}] {{ t.errorMessage }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCw as RefreshCwIcon } from 'lucide-vue-next'
import { listTasks, cancelTask } from '../api/tasks'
import type { Task } from '../types'

const tasks = ref<Task[]>([])
const loading = ref(false)

function getStatusLabel(status: Task['status']) {
  switch (status) {
    case 'queued': return '排队中'
    case 'running': return '执行中'
    case 'cancelling': return '取消中'
    case 'succeeded': return '已完成'
    case 'failed': return '失败'
    case 'cancelled': return '已取消'
    default: return status
  }
}

function getStatusBadgeClass(status: Task['status']) {
  switch (status) {
    case 'succeeded': return 'badge-success'
    case 'running': return 'badge-primary'
    case 'failed': return 'badge-danger'
    case 'cancelling':
    case 'cancelled': return 'badge-warning'
    default: return 'badge-neutral'
  }
}

function formatDate(iso: string) {
  if (!iso) return ''
  return new Date(iso).toLocaleString()
}

async function loadTasks() {
  loading.value = true
  try {
    tasks.value = await listTasks()
  } finally {
    loading.value = false
  }
}

async function handleCancel(id: number) {
  try {
    await cancelTask(id)
    await loadTasks()
  } catch (err: any) {
    alert(err?.message || '取消任务失败')
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.tasks-page {
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
  padding: 20px;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.task-card {
  padding: 16px;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.task-type-badge {
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.task-id {
  color: var(--text-dim);
  font-size: 0.8rem;
}

.task-msg {
  font-size: 0.875rem;
  color: var(--text-main);
  font-weight: 500;
}

.task-status-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-bar-wrap {
  width: 100%;
  height: 6px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366f1, #06b6d4);
  border-radius: 999px;
  transition: width 0.3s ease;
}

.task-footer-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-error-box {
  background: rgba(244, 63, 94, 0.12);
  border: 1px solid rgba(244, 63, 94, 0.25);
  color: #fb7185;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.text-xs {
  font-size: 0.725rem;
}

.empty-state {
  text-align: center;
  color: var(--text-dim);
  padding: 40px;
  font-size: 0.875rem;
}
</style>
