<template>
  <div class="app-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-icon">
          <LayersIcon class="icon-inner" />
        </div>
        <div class="brand-info">
          <span class="brand-name">DevEnv</span>
          <span class="brand-sub font-mono">v1.0.0</span>
        </div>
      </div>

      <nav class="nav-menu">
        <router-link to="/" class="nav-item" active-class="active">
          <LayoutDashboardIcon class="nav-icon" />
          <span>控制台 (Dashboard)</span>
        </router-link>

        <router-link to="/projects" class="nav-item" active-class="active">
          <FolderIcon class="nav-icon" />
          <span>项目管理 (Projects)</span>
        </router-link>

        <router-link to="/php" class="nav-item" active-class="active">
          <CodeIcon class="nav-icon" />
          <span>PHP 环境 (PHP)</span>
        </router-link>

        <router-link to="/services" class="nav-item" active-class="active">
          <ServerIcon class="nav-icon" />
          <span>服务实例 (Services)</span>
        </router-link>

        <router-link to="/hosts" class="nav-item" active-class="active">
          <GlobeIcon class="nav-icon" />
          <span>虚拟主机 (Hosts)</span>
        </router-link>

        <router-link to="/tasks" class="nav-item" active-class="active">
          <ListTodoIcon class="nav-icon" />
          <span>任务作业 (Tasks)</span>
        </router-link>

        <router-link to="/settings" class="nav-item" active-class="active">
          <SettingsIcon class="nav-icon" />
          <span>全局设置 (Settings)</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="sys-badge font-mono">
          <span class="status-indicator active"></span>
          <span>SQLite system_ OK</span>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <div class="main-wrapper">
      <header class="topbar">
        <div class="topbar-breadcrumb font-mono">
          <span>DevEnv</span>
          <span class="separator">/</span>
          <span class="curr-page">{{ currentRouteName }}</span>
        </div>

        <div class="topbar-right">
          <span class="badge badge-primary font-mono">Win64 · Go & Wails</span>
        </div>
      </header>

      <main class="content-scroll">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Layers as LayersIcon,
  LayoutDashboard as LayoutDashboardIcon,
  Folder as FolderIcon,
  Code as CodeIcon,
  Server as ServerIcon,
  Globe as GlobeIcon,
  ListTodo as ListTodoIcon,
  Settings as SettingsIcon,
} from 'lucide-vue-next'

const route = useRoute()
const currentRouteName = computed(() => {
  return String(route.name || 'Dashboard')
})
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-app);
  overflow: hidden;
}

/* Sidebar */
.sidebar {
  width: 250px;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  padding: 18px 14px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px 22px 10px;
  border-bottom: 1px solid var(--border-subtle);
}

.brand-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #6366f1 0%, #06b6d4 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-glow);
}

.icon-inner {
  width: 20px;
  height: 20px;
  color: #fff;
}

.brand-info {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 1.15rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #fff;
}

.brand-sub {
  font-size: 0.7rem;
  color: var(--text-dim);
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 16px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.15s ease;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main);
}

.nav-item.active {
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
  font-weight: 600;
}

.nav-icon {
  width: 18px;
  height: 18px;
}

.sidebar-footer {
  margin-top: auto;
  padding-top: 14px;
  border-top: 1px solid var(--border-subtle);
}

.sys-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.725rem;
  color: var(--text-dim);
  background: rgba(15, 23, 42, 0.6);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.status-indicator {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #64748b;
}
.status-indicator.active {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}

/* Main Area */
.main-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.topbar {
  height: 52px;
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(12px);
}

.topbar-breadcrumb {
  font-size: 0.825rem;
  color: var(--text-dim);
  display: flex;
  align-items: center;
  gap: 8px;
}

.separator {
  color: rgba(255, 255, 255, 0.2);
}

.curr-page {
  color: var(--text-main);
  font-weight: 600;
}

.content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
</style>
