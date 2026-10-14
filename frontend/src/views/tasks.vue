<template>
  <h-page title="任务与后台作业中心" desc="查看环境安装、备份、恢复及异步处理任务的进度与状态">
    <template #extra>
      <p-button 
        label="刷新" 
        icon="pi pi-refresh" 
        severity="secondary" 
        outlined 
        :loading="loading" 
        @click="loadTasks" 
      />
    </template>

    <!-- Tasks List -->
    <h-card>
      <div v-if="tasks.length === 0" class="empty-state">
        <span v-if="loading">加载任务中...</span>
        <span v-else>当前没有正在运行或历史记录的任务</span>
      </div>

      <div v-else class="task-items-list">
        <div v-for="t in tasks" :key="t.id" class="task-box">
          <div class="task-top">
            <div class="task-left">
              <span class="task-type font-mono">{{ t.taskType }}</span>
              <span class="task-id font-mono">#{{ t.id }}</span>
              <span class="task-msg">{{ t.message || '处理中...' }}</span>
            </div>

            <div class="task-right">
              <p-tag 
                :severity="getStatusSeverity(t.status)" 
                :value="getStatusLabel(t.status)" 
              />
              <p-button 
                v-if="t.status === 'running' || t.status === 'queued'" 
                label="取消任务" 
                severity="danger" 
                size="small" 
                text 
                @click="handleCancel(t.id)" 
              />
            </div>
          </div>

          <p-progress-bar :value="t.progress" :showValue="false" style="height: 6px;" />

          <div class="task-meta-line">
            <span class="text-sm text-muted">创建时间: {{ formatDate(t.createdAt) }}</span>
            <span class="text-sm font-mono">{{ t.progress }}%</span>
          </div>

          <div v-if="t.errorMessage" class="task-error-box font-mono">
            错误: [{{ t.errorCode }}] {{ t.errorMessage }}
          </div>
        </div>
      </div>
    </h-card>
  </h-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import HPage from '@/components/h-page/h-page.vue'
import HCard from '@/components/h-card/h-card.vue'
import { listTasks, cancelTask } from '@/api/tasks'
import type { Task } from '@/types'

defineOptions({
  name: 'Tasks'
})

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

function getStatusSeverity(status: Task['status']) {
  switch (status) {
    case 'succeeded': return 'success'
    case 'running': return 'info'
    case 'failed': return 'danger'
    case 'cancelling':
    case 'cancelled': return 'warn'
    default: return 'secondary'
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
    toast.info('任务已请求取消')
    await loadTasks()
  } catch (err: any) {
    toast.error(err?.message || '取消任务失败')
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
}

.tasks-card {
  padding: 20px;
}

.task-items-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.task-box {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.task-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.task-type {
  background: #eff6ff;
  color: #2563eb;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  border: 1px solid #bfdbfe;
}

.task-id {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.task-msg {
  font-size: 0.9rem;
  font-weight: 400;
  color: var(--text-main);
}

.task-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.task-meta-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-error-box {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #ef4444;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-size: 0.775rem;
}

.text-sm { font-size: 0.8rem; }
.text-muted { color: #64748b; }

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
}
</style>
