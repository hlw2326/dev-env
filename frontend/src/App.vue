<template>
  <div class="app-shell">
    <Toast position="top-right" />

    <!-- 自定义窗口头部 / 标题栏 (Custom Header) -->
    <header class="custom-titlebar">
      <!-- 左侧：品牌与模块信息 -->
      <div class="titlebar-left">
        <div class="brand-mini-logo">
          <i class="pi pi-box brand-icon"></i>
        </div>
        <span class="brand-name">DevEnv</span>
        <span class="brand-sep">/</span>
        <span class="page-name font-mono">{{ currentRouteName }}</span>
      </div>

      <!-- 中间：可拖拽区域与标题提示 (双击可最大化/还原) -->
      <div class="titlebar-center" @dblclick="handleToggleMaximize">
        <span class="titlebar-title">本地开发环境管理平台</span>
      </div>

      <!-- 右侧：环境信息与自定义窗口控制按钮 -->
      <div class="titlebar-right">
        <span class="env-tag font-mono">Win64 · Go & Wails</span>

        <div class="window-controls">
          <button
            type="button"
            class="win-btn win-btn-min"
            title="最小化"
            @click="handleMinimize"
          >
            <i class="pi pi-minus"></i>
          </button>
          <button
            type="button"
            class="win-btn win-btn-max"
            :title="isMaximized ? '还原' : '最大化'"
            @click="handleToggleMaximize"
          >
            <i :class="isMaximized ? 'pi pi-clone' : 'pi pi-stop'"></i>
          </button>
          <button
            type="button"
            class="win-btn win-btn-close"
            title="关闭"
            @click="handleClose"
          >
            <i class="pi pi-times"></i>
          </button>
        </div>
      </div>
    </header>

    <!-- 主布局：侧边栏 + 业务内容区 -->
    <div class="app-body">
      <!-- 纯净亮色侧边栏 -->
      <aside class="sidebar">
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
            <span>SQLite system_ 正常</span>
          </div>
        </div>
      </aside>

      <!-- 视图滚动内容展示区 -->
      <main class="content-scroll">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import Toast from 'primevue/toast'
import {
  WindowMinimise,
  WindowToggleMaximise,
  WindowIsMaximised,
  Quit
} from '../wailsjs/runtime/runtime'

const route = useRoute()
const currentRouteName = computed(() => {
  return String(route.name || '控制台')
})

const isMaximized = ref(false)

async function updateMaximizedState() {
  try {
    if (typeof (window as any)?.runtime?.WindowIsMaximised === 'function') {
      isMaximized.value = await WindowIsMaximised()
    }
  } catch {
    // 浏览器独立预览环境静默跳过
  }
}

function handleMinimize() {
  try {
    WindowMinimise()
  } catch {
    // 浏览器环境兜底
  }
}

function handleToggleMaximize() {
  try {
    WindowToggleMaximise()
    setTimeout(updateMaximizedState, 150)
  } catch {
    // 浏览器环境兜底
  }
}

function handleClose() {
  try {
    Quit()
  } catch {
    // 浏览器环境兜底
  }
}

onMounted(() => {
  updateMaximizedState()
  window.addEventListener('resize', updateMaximizedState)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateMaximizedState)
})
</script>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-app);
  overflow: hidden;
}

/* 自定义窗口标题栏 */
.custom-titlebar {
  height: 38px;
  background: #ffffff;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0 0 14px;
  user-select: none;
  --wails-draggable: drag;
  flex-shrink: 0;
  z-index: 1000;
}

.titlebar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  --wails-draggable: drag;
}

.brand-mini-logo {
  width: 22px;
  height: 22px;
  border-radius: var(--radius-sm);
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 0.75rem;
}

.brand-icon {
  font-size: 0.8rem;
}

.brand-name {
  font-size: 0.875rem;
  font-weight: 400;
  color: var(--text-main);
  letter-spacing: -0.01em;
}

.brand-sep {
  color: #cbd5e1;
  font-size: 0.8rem;
}

.page-name {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.titlebar-center {
  flex: 1;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  --wails-draggable: drag;
  cursor: default;
}

.titlebar-title {
  font-size: 0.775rem;
  color: #94a3b8;
  letter-spacing: 0.02em;
}

.titlebar-right {
  display: flex;
  align-items: center;
  height: 100%;
  gap: 12px;
}

.env-tag {
  font-size: 0.725rem;
  color: #64748b;
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid #e2e8f0;
  --wails-draggable: no-drag;
}

.window-controls {
  display: flex;
  height: 100%;
  --wails-draggable: no-drag;
}

.win-btn {
  width: 44px;
  height: 100%;
  border: none;
  background: transparent;
  color: #475569;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  outline: none;
  transition: background-color 0.15s, color 0.15s;
  font-size: 0.725rem;
}

.win-btn:hover {
  background-color: #f1f5f9;
  color: #0f172a;
}

.win-btn-close:hover {
  background-color: #ef4444 !important;
  color: #ffffff !important;
}

.win-btn-close:active {
  background-color: #dc2626 !important;
  color: #ffffff !important;
}

/* 主体区域 */
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 侧边导航栏 */
.sidebar {
  width: 210px;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 12px 8px;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-md);
  color: #475569;
  text-decoration: none;
  font-size: 0.85rem;
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
  font-size: 0.95rem;
}

.sidebar-footer {
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--border-color);
}

.sys-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.725rem;
  color: var(--text-muted);
  background: #f8fafc;
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #94a3b8;
}
.status-dot.online {
  background: #10b981;
}

/* 页面内容区 */
.content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 18px 22px;
}
</style>
