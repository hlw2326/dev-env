<template>
  <div class="app-shell">
    <Toast position="top-right" />

    <!-- 自定义窗口头部 / 标题栏 (Custom Header) -->
    <header class="custom-titlebar">
      <!-- 左侧：品牌与模块面包屑 -->
      <div class="titlebar-left">
        <div class="brand-mini-logo">
          <i class="pi pi-box brand-icon"></i>
        </div>
        <span class="brand-name">DevEnv</span>
        <span class="brand-sep">/</span>
        <span class="page-name font-mono">{{ currentRouteName }}</span>
      </div>

      <!-- 中间：可拖拽区域与居中提示 (双击可最大化/还原) -->
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
      <!-- 极简 Slate 亮色侧边栏 -->
      <aside class="sidebar">
        <nav class="nav-menu">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="nav-item"
            active-class="active"
          >
            <i :class="[item.icon, 'nav-icon']"></i>
            <span>{{ item.label }}</span>
            <span v-if="item.badge" class="nav-badge font-mono">{{ item.badge }}</span>
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

interface NavItem {
  label: string
  icon: string
  to: string
  badge?: string | number
}

const navItems: NavItem[] = [
  { label: '控制台', icon: 'pi pi-th-large', to: '/' },
  { label: '项目管理', icon: 'pi pi-folder', to: '/projects' },
  { label: 'PHP 环境', icon: 'pi pi-code', to: '/php' },
  { label: '服务实例', icon: 'pi pi-server', to: '/services' },
  { label: '虚拟主机', icon: 'pi pi-globe', to: '/hosts' },
  { label: '任务作业', icon: 'pi pi-list-check', to: '/tasks' },
  { label: '全局设置', icon: 'pi pi-cog', to: '/settings' }
]

const route = useRoute()
const currentRouteName = computed(() => {
  return String(route.meta.title || route.name || '控制台')
})

const isMaximized = ref(false)

async function updateMaximizedState() {
  try {
    if (typeof (window as any)?.runtime?.WindowIsMaximised === 'function') {
      isMaximized.value = await WindowIsMaximised()
    }
  } catch {
    // 浏览器独立预览环境跳过
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

<style scoped lang="scss">
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
  background: var(--bg-card);
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
  background: var(--p-primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: var(--font-xs);
}

.brand-icon {
  font-size: 0.8rem;
}

.brand-name {
  font-size: var(--font-base);
  font-weight: 500;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.brand-sep {
  color: var(--border-hover);
  font-size: var(--font-sm);
}

.page-name {
  font-size: var(--font-sm);
  color: var(--text-secondary);
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
  font-size: var(--font-xs);
  color: var(--text-muted);
  letter-spacing: 0.02em;
}

.titlebar-right {
  display: flex;
  align-items: center;
  height: 100%;
  gap: 12px;
}

.env-tag {
  font-size: var(--font-xs);
  color: var(--text-secondary);
  background: var(--bg-app);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
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
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  outline: none;
  transition: background-color 0.15s, color 0.15s;
  font-size: var(--font-xs);
}

.win-btn:hover {
  background-color: var(--bg-app);
  color: var(--text-primary);
}

.win-btn-close:hover {
  background-color: var(--danger-color) !important;
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
  gap: 4px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-decoration: none;
  font-size: var(--font-sm);
  font-weight: 400;
  transition: all 0.15s ease;
}

.nav-item:hover {
  background: var(--bg-app);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--bg-app);
  color: var(--text-primary);
  border-left: 3px solid var(--p-primary-color);
  border-top-left-radius: 2px;
  border-bottom-left-radius: 2px;
}

.nav-icon {
  font-size: var(--font-md);
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
  font-size: var(--font-xs);
  color: var(--text-muted);
  background: var(--bg-app);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
}
.status-dot.online {
  background: var(--success-color);
}

/* 页面内容区 */
.content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: var(--h-padding) 20px;
}
</style>
