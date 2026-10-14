<template>
  <div class="dashboard-page fade-in">
    <!-- Top Greeting Banner -->
    <div class="welcome-header glass-panel">
      <div class="welcome-text">
        <h2>DevEnv 本地开发环境</h2>
        <p>Windows 现代高效 PHP / Web 开发栈一站式控制中心</p>
      </div>
      <div class="quick-status-pills">
        <span class="status-pill">
          <span class="dot online"></span>
          SQLite 数据库已连接 (system_ 架构)
        </span>
        <span class="status-pill">
          <span class="dot online"></span>
          进程生命周期管理器运行中
        </span>
      </div>
    </div>

    <!-- Metric Stat Cards -->
    <div class="stats-grid">
      <div class="stat-card glass-panel" @click="$router.push('/projects')">
        <div class="stat-icon-wrap indigo">
          <FolderIcon class="stat-icon" />
        </div>
        <div class="stat-data">
          <span class="stat-val">{{ projects.length }}</span>
          <span class="stat-title">已登记项目</span>
        </div>
      </div>

      <div class="stat-card glass-panel" @click="$router.push('/php')">
        <div class="stat-icon-wrap cyan">
          <CodeIcon class="stat-icon" />
        </div>
        <div class="stat-data">
          <span class="stat-val">{{ phpVersions.length }}</span>
          <span class="stat-title">PHP 运行环境</span>
        </div>
      </div>

      <div class="stat-card glass-panel" @click="$router.push('/services')">
        <div class="stat-icon-wrap emerald">
          <ServerIcon class="stat-icon" />
        </div>
        <div class="stat-data">
          <span class="stat-val">{{ runningServicesCount }} / {{ services.length }}</span>
          <span class="stat-title">运行中服务</span>
        </div>
      </div>

      <div class="stat-card glass-panel" @click="$router.push('/hosts')">
        <div class="stat-icon-wrap amber">
          <GlobeIcon class="stat-icon" />
        </div>
        <div class="stat-data">
          <span class="stat-val">{{ virtualHosts.length }}</span>
          <span class="stat-title">虚拟主机 / 域名</span>
        </div>
      </div>
    </div>

    <!-- Quick Services Overview & Recent Projects -->
    <div class="content-split-grid">
      <!-- Left: Services Quick Control -->
      <div class="panel-section glass-panel">
        <div class="section-title-row">
          <h3>
            <ActivityIcon class="title-icon" />
            核心服务快捷控制
          </h3>
          <button class="btn btn-secondary btn-sm" @click="loadData">
            <RefreshCwIcon class="btn-icon" :class="{ 'spin-icon': loading }" />
            刷新
          </button>
        </div>

        <div v-if="services.length === 0" class="empty-state">
          暂无已配置的服务，点击上方服务管理添加
        </div>

        <div v-else class="services-quick-list">
          <div v-for="srv in services" :key="srv.id" class="service-quick-item">
            <div class="srv-left">
              <span class="status-indicator" :class="srv.status === 'running' ? 'active' : ''"></span>
              <div class="srv-info">
                <span class="srv-name">{{ srv.name }}</span>
                <span class="srv-sub">端口: {{ srv.port || '无' }} · {{ srv.status === 'running' ? `PID: ${srv.pid}` : '已停止' }}</span>
              </div>
            </div>

            <div class="srv-actions">
              <button 
                v-if="srv.status !== 'running'" 
                class="btn btn-success btn-sm"
                :disabled="operatingId === srv.id"
                @click="handleStart(srv.id)"
              >
                启动
              </button>
              <button 
                v-else 
                class="btn btn-danger btn-sm"
                :disabled="operatingId === srv.id"
                @click="handleStop(srv.id)"
              >
                停止
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Recent Projects -->
      <div class="panel-section glass-panel">
        <div class="section-title-row">
          <h3>
            <LayersIcon class="title-icon" />
            项目列表
          </h3>
          <button class="btn btn-primary btn-sm" @click="$router.push('/projects')">
            进入项目管理
          </button>
        </div>

        <div v-if="projects.length === 0" class="empty-state">
          尚未登记任何项目，进入项目管理快速导入或创建
        </div>

        <div v-else class="project-quick-list">
          <div v-for="proj in projects.slice(0, 5)" :key="proj.id" class="project-quick-item">
            <div class="proj-meta">
              <span class="proj-title">{{ proj.name }}</span>
              <span class="proj-path">{{ proj.path }}</span>
            </div>
            <span class="badge" :class="proj.status === 'ready' ? 'badge-success' : 'badge-danger'">
              {{ proj.status === 'ready' ? '就绪' : '缺失' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  Folder as FolderIcon, 
  Code as CodeIcon, 
  Server as ServerIcon, 
  Globe as GlobeIcon, 
  Activity as ActivityIcon, 
  RefreshCw as RefreshCwIcon,
  Layers as LayersIcon 
} from 'lucide-vue-next'
import { listProjects } from '../api/projects'
import { listVersions } from '../api/php'
import { listServices, startService, stopService } from '../api/services'
import { listVirtualHosts } from '../api/hosts'
import type { Project, PHPVersion, ServiceInstance, VirtualHost } from '../types'

const projects = ref<Project[]>([])
const phpVersions = ref<PHPVersion[]>([])
const services = ref<ServiceInstance[]>([])
const virtualHosts = ref<VirtualHost[]>([])
const loading = ref(false)
const operatingId = ref<number | null>(null)

const runningServicesCount = computed(() => {
  return services.value.filter(s => s.status === 'running').length
})

async function loadData() {
  loading.value = true
  try {
    const [p, v, s, h] = await Promise.all([
      listProjects().catch(() => []),
      listVersions().catch(() => []),
      listServices().catch(() => []),
      listVirtualHosts().catch(() => []),
    ])
    projects.value = p
    phpVersions.value = v
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
    await loadData()
  } catch (err: any) {
    alert(err?.message || '启动失败')
  } finally {
    operatingId.value = null
  }
}

async function handleStop(id: number) {
  operatingId.value = id
  try {
    await stopService(id)
    await loadData()
  } catch (err: any) {
    alert(err?.message || '停止失败')
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
  gap: 20px;
}

.welcome-header {
  padding: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.8) 0%, rgba(15, 23, 42, 0.85) 100%);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.welcome-text h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #fff;
  letter-spacing: -0.02em;
}

.welcome-text p {
  color: var(--text-muted);
  font-size: 0.875rem;
  margin-top: 4px;
}

.quick-status-pills {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(15, 23, 42, 0.6);
  padding: 5px 12px;
  border-radius: var(--radius-full);
  font-size: 0.775rem;
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}
.dot.online {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  padding: 18px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.stat-card:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.2);
  background: var(--bg-card-hover);
}

.stat-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon-wrap.indigo { background: rgba(99, 102, 241, 0.15); color: #818cf8; }
.stat-icon-wrap.cyan { background: rgba(6, 182, 212, 0.15); color: #22d3ee; }
.stat-icon-wrap.emerald { background: rgba(16, 185, 129, 0.15); color: #34d399; }
.stat-icon-wrap.amber { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }

.stat-icon {
  width: 24px;
  height: 24px;
}

.stat-data {
  display: flex;
  flex-direction: column;
}
.stat-val {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.2;
}
.stat-title {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.content-split-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.panel-section {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title-row h3 {
  font-size: 1.05rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-icon {
  width: 18px;
  height: 18px;
  color: var(--accent-cyan);
}

.btn-icon {
  width: 14px;
  height: 14px;
}

.services-quick-list, .project-quick-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.service-quick-item, .project-quick-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  transition: all 0.15s;
}
.service-quick-item:hover, .project-quick-item:hover {
  background: rgba(15, 23, 42, 0.8);
  border-color: rgba(255, 255, 255, 0.12);
}

.srv-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #64748b;
}
.status-indicator.active {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}

.srv-info, .proj-meta {
  display: flex;
  flex-direction: column;
}

.srv-name, .proj-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-main);
}

.srv-sub, .proj-path {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.proj-path {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-state {
  text-align: center;
  color: var(--text-dim);
  padding: 30px;
  font-size: 0.875rem;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
