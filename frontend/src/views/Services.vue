<template>
  <div class="services-page">
    <!-- Header -->
    <div class="page-header-bar">
      <div>
        <h1 class="page-title">服务实例管理</h1>
        <p class="page-subtitle">监控与管理本地 Web 服务器 (Nginx)、数据库 (MySQL) 与 PHP FastCGI 进程</p>
      </div>
      <div class="header-actions">
        <Button 
          label="刷新状态" 
          icon="pi pi-refresh" 
          severity="secondary" 
          outlined 
          :loading="loading" 
          @click="loadServices" 
        />
      </div>
    </div>

    <!-- Services Grid -->
    <div class="services-grid">
      <div 
        v-for="srv in services" 
        :key="srv.id" 
        class="service-card surface-card"
        :class="{ 'is-running': srv.status === 'running' }"
      >
        <div class="srv-header">
          <div class="srv-badge-title">
            <span class="type-pill" :class="srv.serviceType">
              {{ srv.serviceType.toUpperCase() }}
            </span>
            <span class="srv-name">{{ srv.name }}</span>
          </div>

          <Tag 
            :severity="srv.status === 'running' ? 'success' : 'secondary'"
            :value="srv.status === 'running' ? '运行中 (Running)' : '已停止 (Stopped)'" 
            :icon="srv.status === 'running' ? 'pi pi-check-circle' : 'pi pi-minus-circle'"
          />
        </div>

        <div class="srv-metrics-grid">
          <div class="metric-col">
            <span class="m-lbl">监听端口</span>
            <span class="m-val font-mono">{{ srv.port ? srv.port : '无' }}</span>
          </div>
          <div class="metric-col">
            <span class="m-lbl">进程 PID</span>
            <span class="m-val font-mono">{{ srv.pid ? srv.pid : '-' }}</span>
          </div>
          <div class="metric-col">
            <span class="m-lbl">开机自启</span>
            <span class="m-val">{{ srv.autoStart ? '开启' : '关闭' }}</span>
          </div>
        </div>

        <div class="srv-footer-actions">
          <Button 
            v-if="srv.status !== 'running'"
            label="启动服务" 
            icon="pi pi-play" 
            severity="success" 
            class="flex-1"
            :loading="operatingId === srv.id"
            @click="handleStart(srv.id)"
          />

          <template v-else>
            <Button 
              label="停止" 
              icon="pi pi-stop" 
              severity="danger" 
              outlined
              class="flex-1"
              :loading="operatingId === srv.id"
              @click="handleStop(srv.id)"
            />
            <Button 
              label="重启" 
              icon="pi pi-replay" 
              severity="secondary" 
              outlined
              class="flex-1"
              :loading="operatingId === srv.id"
              @click="handleRestart(srv.id)"
            />
          </template>

          <Button 
            icon="pi pi-terminal" 
            label="日志" 
            severity="secondary" 
            text 
            @click="openLogs(srv)" 
          />
        </div>
      </div>
    </div>

    <!-- Log Console Panel -->
    <div v-if="activeLogService" class="log-panel surface-card">
      <div class="log-panel-header">
        <div class="log-panel-title">
          <i class="pi pi-terminal"></i>
          <span>{{ activeLogService.name }} - 控制台实时日志</span>
        </div>
        <div class="log-panel-actions">
          <Button label="刷新日志" icon="pi pi-refresh" severity="secondary" text size="small" @click="fetchLogs" />
          <Button label="关闭" severity="secondary" text size="small" @click="activeLogService = null" />
        </div>
      </div>

      <div class="log-console font-mono">
        <div v-if="logs.length === 0" class="log-empty">暂无输出日志</div>
        <div v-for="(line, idx) in logs" :key="idx" class="log-row">
          <span class="log-num">{{ idx + 1 }}</span>
          <span class="log-content">{{ line }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { listServices, startService, stopService, restartService, getServiceLogs } from '../api/services'
import type { ServiceInstance } from '../types'

const toast = useToast()
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
    toast.add({ severity: 'success', summary: '成功', detail: '服务启动成功', life: 2500 })
    await loadServices()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: '错误', detail: err?.message || '启动服务失败', life: 3500 })
  } finally {
    operatingId.value = null
  }
}

async function handleStop(id: number) {
  operatingId.value = id
  try {
    await stopService(id)
    toast.add({ severity: 'info', summary: '已停止', detail: '服务已停止', life: 2500 })
    await loadServices()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: '错误', detail: err?.message || '停止服务失败', life: 3500 })
  } finally {
    operatingId.value = null
  }
}

async function handleRestart(id: number) {
  operatingId.value = id
  try {
    await restartService(id)
    toast.add({ severity: 'success', summary: '成功', detail: '服务重启成功', life: 2500 })
    await loadServices()
  } catch (err: any) {
    toast.add({ severity: 'error', summary: '错误', detail: err?.message || '重启服务失败', life: 3500 })
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

.services-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 16px;
}

.service-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.service-card.is-running {
  border-color: #86efac;
}

.srv-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.srv-badge-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.type-pill {
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 0.725rem;
  font-weight: 700;
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}
.type-pill.php { background: #eff6ff; color: #2563eb; border-color: #bfdbfe; }
.type-pill.nginx { background: #ecfdf5; color: #059669; border-color: #a7f3d0; }
.type-pill.mysql { background: #fffbeb; color: #d97706; border-color: #fde68a; }

.srv-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.srv-metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  background: #f8fafc;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  border: 1px solid #e2e8f0;
}

.metric-col {
  display: flex;
  flex-direction: column;
}

.m-lbl {
  font-size: 0.725rem;
  color: var(--text-muted);
}

.m-val {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  margin-top: 2px;
}

.srv-footer-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
}

.flex-1 {
  flex: 1;
}

.log-panel {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.log-panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main);
}

.log-console {
  background: #0f172a;
  border-radius: var(--radius-md);
  padding: 14px;
  height: 240px;
  overflow-y: auto;
  color: #f8fafc;
}

.log-row {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
  line-height: 1.5;
}

.log-num {
  color: #64748b;
  min-width: 24px;
  user-select: none;
}

.log-content {
  color: #cbd5e1;
  word-break: break-all;
}

.log-empty {
  text-align: center;
  padding: 70px 0;
  color: #64748b;
  font-size: 0.85rem;
}
</style>
