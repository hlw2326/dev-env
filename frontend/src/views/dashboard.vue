<template>
  <h-page title="开发环境概览" desc="Windows 本地高效 PHP / Web 开发栈一站式控制中心">
    <template #extra>
      <p-tag severity="success" value="SQLite: system_ 正常" icon="pi pi-check" />
      <p-tag severity="info" value="进程管理器: 运行中" icon="pi pi-server" />
    </template>

    <!-- Stats Grid -->
    <div class="stats-row">
      <div class="stat-box surface-card" @click="$router.push('/projects')">
        <div class="stat-icon-bg bg-blue">
          <i class="pi pi-folder"></i>
        </div>
        <div class="stat-content">
          <span class="stat-number">{{ projects.length }}</span>
          <span class="stat-label">已登记项目</span>
        </div>
      </div>

      <div class="stat-box surface-card" @click="$router.push('/runtimes')">
        <div class="stat-icon-bg bg-indigo">
          <i class="pi pi-box"></i>
        </div>
        <div class="stat-content">
          <span class="stat-number">{{ runtimesCount }}</span>
          <span class="stat-label">运行环境</span>
        </div>
      </div>

      <div class="stat-box surface-card" @click="$router.push('/services')">
        <div class="stat-icon-bg bg-emerald">
          <i class="pi pi-server"></i>
        </div>
        <div class="stat-content">
          <span class="stat-number">{{ runningServicesCount }} / {{ services.length }}</span>
          <span class="stat-label">运行中服务</span>
        </div>
      </div>

      <div class="stat-box surface-card" @click="$router.push('/hosts')">
        <div class="stat-icon-bg bg-amber">
          <i class="pi pi-globe"></i>
        </div>
        <div class="stat-content">
          <span class="stat-number">{{ virtualHosts.length }}</span>
          <span class="stat-label">虚拟主机 / 域名</span>
        </div>
      </div>
    </div>

    <!-- Quick Services & Projects Grid -->
    <div class="panels-grid">
      <!-- Services Quick List -->
      <div class="panel surface-card">
        <div class="panel-header">
          <div class="panel-title">
            <i class="pi pi-bolt text-blue"></i>
            <span>核心服务状态与控制</span>
          </div>
          <p-button 
            icon="pi pi-refresh" 
            label="刷新" 
            size="small" 
            severity="secondary" 
            outlined 
            :loading="loading" 
            @click="loadData" 
          />
        </div>

        <div v-if="services.length === 0" class="empty-block">
          暂无已配置的服务
        </div>

        <div v-else class="services-list">
          <div v-for="srv in services" :key="srv.id" class="service-row">
            <div class="service-info">
              <div class="service-title-line">
                <span class="service-name">{{ srv.name }}</span>
                <p-tag 
                  :severity="srv.status === 'running' ? 'success' : 'secondary'"
                  :value="srv.status === 'running' ? '运行中' : '已停止'" 
                  size="small"
                />
              </div>
              <span class="service-meta font-mono">
                端口: {{ srv.port || '无' }} · {{ srv.status === 'running' ? `PID: ${srv.pid}` : '空闲' }}
              </span>
            </div>

            <div class="service-actions">
              <p-button 
                v-if="srv.status !== 'running'"
                label="启动" 
                icon="pi pi-play" 
                size="small" 
                severity="success"
                :loading="operatingId === srv.id"
                @click="handleStart(srv.id)"
              />
              <p-button 
                v-else
                label="停止" 
                icon="pi pi-stop" 
                size="small" 
                severity="danger" 
                outlined
                :loading="operatingId === srv.id"
                @click="handleStop(srv.id)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Projects -->
      <div class="panel surface-card">
        <div class="panel-header">
          <div class="panel-title">
            <i class="pi pi-list text-indigo"></i>
            <span>最近项目</span>
          </div>
          <p-button 
            label="进入项目管理" 
            icon="pi pi-arrow-right" 
            iconPos="right" 
            size="small" 
            @click="$router.push('/projects')" 
          />
        </div>

        <div v-if="projects.length === 0" class="empty-block">
          尚未登记任何项目
        </div>

        <div v-else class="projects-list">
          <div v-for="proj in projects.slice(0, 5)" :key="proj.id" class="project-row">
            <div class="project-info">
              <span class="project-name">{{ proj.name }}</span>
              <span class="project-path font-mono" :title="proj.path">{{ proj.path }}</span>
            </div>
            <p-tag 
              :severity="proj.status === 'ready' ? 'success' : 'danger'"
              :value="proj.status === 'ready' ? '就绪' : '缺失'" 
              size="small"
            />
          </div>
        </div>
      </div>
    </div>
  </h-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import HPage from '@/components/h-page/h-page.vue'
import { listProjects } from '@/api/projects'
import { listAllRuntimes } from '@/api/runtimes'
import { listServices, startService, stopService } from '@/api/services'
import { listVirtualHosts } from '@/api/hosts'
import type { Project, RuntimeVersion, ServiceInstance, VirtualHost } from '@/types'

defineOptions({
  name: 'Dashboard'
})

const projects = ref<Project[]>([])
const runtimes = ref<RuntimeVersion[]>([])
const services = ref<ServiceInstance[]>([])
const virtualHosts = ref<VirtualHost[]>([])
const loading = ref(false)
const operatingId = ref<number | null>(null)

const runtimesCount = computed(() => runtimes.value.length)
const runningServicesCount = computed(() => {
  return services.value.filter(s => s.status === 'running').length
})

async function loadData() {
  loading.value = true
  try {
    const [p, r, s, h] = await Promise.all([
      listProjects().catch(() => []),
      listAllRuntimes().catch(() => []),
      listServices().catch(() => []),
      listVirtualHosts().catch(() => []),
    ])
    projects.value = p
    runtimes.value = r
    services.value = s
    virtualHosts.value = h
  } finally {
    loading.value = false
  }
}

async function handleStart(id: number) {
  operatingId.value = id
  try {
    await startService(id)
    toast.success('服务启动成功')
    await loadData()
  } catch (err: any) {
    toast.error(err?.message || '服务启动失败')
  } finally {
    operatingId.value = null
  }
}

async function handleStop(id: number) {
  operatingId.value = id
  try {
    await stopService(id)
    toast.info('服务已成功停止')
    await loadData()
  } catch (err: any) {
    toast.error(err?.message || '服务停止失败')
  } finally {
    operatingId.value = null
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: var(--h-padding);
}

.header-status-group {
  display: flex;
  gap: 8px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--h-padding);
  margin-bottom: 4px;
}

.stat-box {
  padding: var(--h-padding);
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  background: var(--bg-card);
}

.stat-box:hover {
  background: #f8fafc;
  border-color: var(--border-hover);
}

.stat-icon-bg {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  flex-shrink: 0;
}
.bg-blue { background: #eff6ff; color: #2563eb; }
.bg-indigo { background: #eef2ff; color: #4f46e5; }
.bg-emerald { background: #ecfdf5; color: #059669; }
.bg-amber { background: #fffbeb; color: #d97706; }

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-number {
  font-size: 1.4rem;
  font-weight: 400;
  color: var(--text-main);
  line-height: 1.2;
}

.stat-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.panels-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--h-padding);
}

.panel {
  padding: var(--h-padding);
  display: flex;
  flex-direction: column;
  gap: var(--h-padding);
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  background: var(--bg-card);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1rem;
  font-weight: 400;
  color: var(--text-main);
}
.text-blue { color: #2563eb; }
.text-indigo { color: #4f46e5; }

.services-list, .projects-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.service-row, .project-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  background: #f8fafc;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.service-row:hover, .project-row:hover {
  background: #ffffff;
  border-color: var(--border-hover);
}

.service-info, .project-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.service-title-line {
  display: flex;
  align-items: center;
  gap: 8px;
}

.service-name, .project-name {
  font-size: 0.9rem;
  font-weight: 400;
  color: var(--text-main);
}

.service-meta, .project-path {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.project-path {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-block {
  text-align: center;
  padding: 32px;
  color: var(--text-dim);
  font-size: 0.85rem;
  background: #f8fafc;
  border: 1px dashed var(--border-color);
  border-radius: var(--h-radius);
}
</style>
