<template>
  <div class="app-layout">
    <Toast position="top-right" />

    <!-- Clean Light Sidebar -->
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-logo">
          <i class="pi pi-box brand-icon"></i>
        </div>
        <div class="brand-text">
          <span class="brand-name">DevEnv</span>
          <span class="brand-badge font-mono">v1.0.0</span>
        </div>
      </div>

      <nav class="nav-menu">
        <router-link to="/" class="nav-item" active-class="active">
          <i class="pi pi-th-large nav-icon"></i>
          <span>控制台</span>
        </router-link>

        <router-link to="/projects" class="nav-item" active-class="active">
          <i class="pi pi-folder nav-icon"></i>
          <span>项目管理</span>
        </router-link>

        <router-link to="/php" class="nav-item" active-class="active">
          <i class="pi pi-code nav-icon"></i>
          <span>PHP 环境</span>
        </router-link>

        <router-link to="/services" class="nav-item" active-class="active">
          <i class="pi pi-server nav-icon"></i>
          <span>服务实例</span>
        </router-link>

        <router-link to="/hosts" class="nav-item" active-class="active">
          <i class="pi pi-globe nav-icon"></i>
          <span>虚拟主机</span>
        </router-link>

        <router-link to="/tasks" class="nav-item" active-class="active">
          <i class="pi pi-list-check nav-icon"></i>
          <span>任务作业</span>
        </router-link>

        <router-link to="/settings" class="nav-item" active-class="active">
          <i class="pi pi-cog nav-icon"></i>
          <span>全局设置</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <div class="sys-badge font-mono">
          <span class="status-dot online"></span>
          <span>SQLite system_ OK</span>
        </div>
      </div>
    </aside>

    <!-- Main Workspace -->
    <div class="main-wrapper">
      <header class="topbar">
        <div class="breadcrumb font-mono">
          <span class="brand-crumb">DevEnv</span>
          <span class="separator">/</span>
          <span class="current-crumb">{{ currentRouteName }}</span>
        </div>

        <div class="topbar-right">
          <span class="env-pill font-mono">Win64 · Go & Wails</span>
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
import Toast from 'primevue/toast'

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
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 16px 12px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px 18px 10px;
  border-bottom: 1px solid var(--border-color);
}

.brand-logo {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-icon {
  font-size: 1.25rem;
  color: #ffffff;
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 1.15rem;
  font-weight: 400;
  color: var(--text-main);
  line-height: 1.2;
}

.brand-badge {
  font-size: 0.7rem;
  color: var(--text-muted);
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 14px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 12px;
  border-radius: var(--radius-md);
  color: #475569;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 400;
  transition: all 0.15s ease;
}

.nav-item:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.nav-item.active {
  background: #eff6ff;
  color: #2563eb;
  font-weight: 400;
}

.nav-icon {
  font-size: 1.05rem;
}

.sidebar-footer {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.sys-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
  color: var(--text-muted);
  background: #f8fafc;
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #94a3b8;
}
.status-dot.online {
  background: #10b981;
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
  background: #ffffff;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.breadcrumb {
  font-size: 0.85rem;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-crumb {
  color: #64748b;
}

.separator {
  color: #cbd5e1;
}

.current-crumb {
  color: var(--text-main);
  font-weight: 400;
}

.env-pill {
  font-size: 0.75rem;
  color: #475569;
  background: #f1f5f9;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  border: 1px solid #e2e8f0;
}

.content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
}
</style>
