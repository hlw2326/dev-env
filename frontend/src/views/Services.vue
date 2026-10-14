<template>
  <div class="services-page fade-in">
    <!-- Header -->
    <div class="page-header glass-panel">
      <div class="header-left">
        <h2>服务实例管理</h2>
        <p>监控与管理本地 Web 服务器 (Nginx)、数据库 (MySQL) 与 PHP FastCGI 等服务进程</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="loadServices">
          <RefreshCwIcon class="btn-icon" :class="{ 'spin-icon': loading }" />
          刷新状态
        </button>
      </div>
    </div>

    <!-- Services Grid -->
    <div class="services-grid">
      <div v-for="srv in services" :key="srv.id" class="service-card glass-panel" :class="{ 'is-running': srv.status === 'running' }">
        <div class="srv-top">
          <div class="srv-title-group">
            <span class="srv-icon-badge" :class="srv.serviceType">
              {{ srv.serviceType.toUpperCase() }}
            </span>
            <div class="srv-titles">
              <span class="srv-card-name">{{ srv.name }}</span>
              <span class="srv-type-tag">类型: {{ srv.serviceType }}</span>
            </div>
          </div>

          <span class="badge" :class="srv.status === 'running' ? 'badge-success' : 'badge-neutral'">
            <span class="dot" :class="srv.status === 'running' ? 'online' : ''"></span>
            {{ srv.status === 'running' ? '运行中 (Running)' : '已停止 (Stopped)' }}
          </span>
        </div>

        <div class="srv-metrics">
          <div class="metric-item">
            <span class="metric-lbl">监听端口:</span>
            <span class="metric-val font-mono">{{ srv.port ? srv.port : '无端口' }}</span>
          </div>
          <div class="metric-item">
            <span class="metric-lbl">进程 PID:</span>
            <span class="metric-val font-mono">{{ srv.pid ? srv.pid : '-' }}</span>
          </div>
          <div class="metric-item">
            <span class="metric-lbl">开机自启:</span>
            <span class="metric-val">{{ srv.autoStart ? '开启' : '关闭' }}</span>
          </div>
        </div>

        <div class="srv-card-actions">
          <button 
            v-if="srv.status !== 'running'" 
            class="btn btn-success btn-sm flex-1"
            :disabled="operatingId === srv.id"
            @click="handleStart(srv.id)"
          >
            <PlayIcon class="btn-icon" />
            启动服务
          </button>

          <template v-else>
            <button 
              class="btn btn-danger btn-sm flex-1"
              :disabled="operatingId === srv.id"
              @click="handleStop(srv.id)"
            >
              <SquareIcon class="btn-icon" />
              停止
            </button>
            <button 
              class="btn btn-secondary btn-sm flex-1"
              :disabled="operatingId === srv.id"
              @click="handleRestart(srv.id)"
            >
              <RotateCwIcon class="btn-icon" />
              重启
            </button>
          </template>

          <button 
            class="btn btn-secondary btn-sm"
            @click="openLogs(srv)"
          >
            <TerminalIcon class="btn-icon" />
            日志
          </button>
        </div>
      </div>
    </div>

    <!-- Live Terminal Console -->
    <div v-if="activeLogService" class="log-console-panel glass-panel">
      <div class="console-header">
        <div class="console-title">
          <TerminalIcon class="console-icon" />
          <span>{{ activeLogService.name }} - 实时输出日志</span>
        </div>
        <div class="console-actions">
          <button class="btn btn-secondary btn-sm" @click="fetchLogs">刷新日志</button>
          <button class="btn btn-secondary btn-sm" @click="activeLogService = null">关闭</button>
        </div>
      </div>

      <div class="console-body font-mono">
        <div v-if="logs.length === 0" class="console-empty">暂无日志输出</div>
        <div v-for="(line, idx) in logs" :key="idx" class="log-line">
          <span class="line-num">{{ idx + 1 }}</span>
          <span class="line-text">{{ line }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { 
  RefreshCw as RefreshCwIcon, 
  Play as PlayIcon, 
  Square as SquareIcon, 
  RotateCw as RotateCwIcon, 
  Terminal as TerminalIcon 
} from 'lucide-vue-next'
import { listServices, startService, stopService, restartService, getServiceLogs } from '../api/services'
import type { ServiceInstance } from '../types'

const services = ref<ServiceInstance[]>([])
const loading = ref(false)
const operatingId = ref<number | null>(null)
const activeLogService = ref<ServiceInstance | null>(null)
const logs = ref<string[]>([])

async function loadServices() {
  loading.value = true
  try {
    services.value = await listServices()
  } finally {
    loading.value = false
  }
}

async function handleStart(id: number) {
  operatingId.value = id
  try {
    await startService(id)
    await loadServices()
  } catch (err: any) {
    alert(err?.message || '启动服务失败')
  } finally {
    operatingId.value = null
  }
}

async function handleStop(id: number) {
  operatingId.value = id
  try {
    await stopService(id)
    await loadServices()
  } catch (err: any) {
    alert(err?.message || '停止服务失败')
  } finally {
    operatingId.value = null
  }
}

async function handleRestart(id: number) {
  operatingId.value = id
  try {
    await restartService(id)
    await loadServices()
  } catch (err: any) {
    alert(err?.message || '重启服务失败')
  } finally {
    operatingId.value = null
  }
}

async function openLogs(srv: ServiceInstance) {
  activeLogService.value = srv
  await fetchLogs()
}

async function fetchLogs() {
  if (!activeLogService.value) return
  try {
    logs.value = await getServiceLogs(activeLogService.value.id, 200)
  } catch {
    logs.value = ['无法加载日志']
  }
}

onMounted(() => {
  loadServices()
})
</script>

<style scoped>
.services-page {
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

.services-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.service-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  border-radius: var(--radius-lg);
  transition: all 0.2s;
}
.service-card.is-running {
  border-color: rgba(16, 185, 129, 0.3);
  box-shadow: 0 0 16px rgba(16, 185, 129, 0.1);
}

.srv-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.srv-title-group {
  display: flex;
  gap: 12px;
  align-items: center;
}

.srv-icon-badge {
  padding: 6px 10px;
  border-radius: var(--radius-md);
  font-size: 0.75rem;
  font-weight: 700;
  font-family: var(--font-mono);
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
}
.srv-icon-badge.nginx { background: rgba(16, 185, 129, 0.15); color: #34d399; }
.srv-icon-badge.mysql { background: rgba(6, 182, 212, 0.15); color: #22d3ee; }

.srv-titles {
  display: flex;
  flex-direction: column;
}

.srv-card-name {
  font-size: 1rem;
  font-weight: 600;
  color: #fff;
}

.srv-type-tag {
  font-size: 0.75rem;
  color: var(--text-dim);
}

.srv-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  background: rgba(15, 23, 42, 0.4);
  padding: 10px 12px;
  border-radius: var(--radius-md);
}

.metric-item {
  display: flex;
  flex-direction: column;
}

.metric-lbl {
  font-size: 0.7rem;
  color: var(--text-dim);
}

.metric-val {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  margin-top: 2px;
}

.srv-card-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
}

.flex-1 {
  flex: 1;
}

.log-console-panel {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #0b0f19;
  border: 1px solid var(--border-card);
}

.console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.console-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--accent-cyan);
}

.console-icon {
  width: 16px;
  height: 16px;
}

.console-body {
  height: 240px;
  overflow-y: auto;
  background: rgba(0, 0, 0, 0.5);
  border-radius: var(--radius-md);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.log-line {
  display: flex;
  gap: 12px;
  font-size: 0.775rem;
  line-height: 1.4;
}

.line-num {
  color: var(--text-dim);
  min-width: 28px;
  user-select: none;
}

.line-text {
  color: #e2e8f0;
  word-break: break-all;
}

.console-empty {
  color: var(--text-dim);
  text-align: center;
  margin-top: 80px;
  font-size: 0.85rem;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #64748b;
}
.dot.online {
  background: #10b981;
  box-shadow: 0 0 6px #10b981;
}
</style>
