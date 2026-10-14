# DevEnv 前端设计与开发规范

> 本规范基于 `zbzs/ele` 项目的成熟实践，定义 DevEnv 前端的设计风格、代码风格与目录结构标准。

---

## 1. 技术栈约束

### 1.1 核心依赖
- **框架**：Vue 3.5+ (Composition API + `<script setup>`)
- **UI 组件库**：PrimeVue 4.5.5
- **主题系统**：`@primevue/themes` 4.5.4（禁用 Aura，使用自定义 Light 主题）
- **状态管理**：Pinia 2.x + pinia-plugin-persistedstate（持久化）
- **路由管理**：Vue Router 4.x（Hash 模式：`createWebHashHistory()`）
- **图标库**：PrimeIcons 8.x
- **构建工具**：Vite 7.x + TypeScript 5.x

### 1.2 开发增效插件（必装）
```typescript
// vite.config.ts 必须配置
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { PrimeVueResolver } from '@primevue/auto-import-resolver'

export default defineConfig({
  plugins: [
    vue(),
    // 自动导入 Vue/Router/Pinia API
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/types/auto-imports.d.ts'
    }),
    // 自动注册组件
    Components({
      resolvers: [PrimeVueResolver()],
      dirs: ['src/components'],
      dts: 'src/types/components.d.ts'
    })
  ]
})
```

---

## 2. 目录结构规范

### 2.1 标准目录树
```
frontend/
├── src/
│   ├── api/              # API 服务层（按业务模块拆分）
│   │   ├── index.ts      # 统一导出
│   │   ├── projects.ts
│   │   ├── php.ts
│   │   ├── services.ts
│   │   └── ...
│   ├── assets/           # 静态资源
│   │   ├── css/
│   │   │   ├── style.css       # 全局样式与设计令牌
│   │   │   └── primevue.css    # PrimeVue 组件样式覆盖
│   │   └── images/
│   ├── bootstrap/        # 应用启动引导
│   │   ├── index.ts      # 引导入口
│   │   └── http.ts       # HTTP 客户端初始化
│   ├── components/       # 全局可复用组件（h- 前缀）
│   │   ├── h-card/
│   │   │   ├── h-card.vue
│   │   │   ├── index.ts
│   │   │   └── types.d.ts
│   │   ├── h-page/
│   │   └── h-layout-header/
│   ├── composables/      # 组合式函数（按业务模块分组）
│   │   ├── index.ts      # 统一导出
│   │   ├── config/
│   │   ├── projects/
│   │   └── services/
│   ├── router/
│   │   └── index.ts      # 路由配置
│   ├── store/            # Pinia 状态管理（按模块拆分）
│   │   ├── index.ts
│   │   ├── config/
│   │   ├── projects/
│   │   └── user/
│   ├── types/            # 全局类型定义
│   │   ├── index.ts
│   │   ├── auto-imports.d.ts    # 自动生成
│   │   └── components.d.ts      # 自动生成
│   ├── utils/            # 工具函数（严禁堆砌，按职责分组）
│   ├── views/            # 页面视图（一级目录即页面）
│   │   ├── main/         # 主应用布局
│   │   │   ├── index.vue
│   │   │   ├── home/
│   │   │   ├── projects/
│   │   │   └── settings/
│   │   └── login/
│   ├── App.vue           # 根组件
│   ├── main.ts           # 应用入口
│   └── style.css         # 快捷导入（实际内容在 assets/css/）
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

### 2.2 关键规则
- **api/** 按业务模块拆分，每个文件对应一个后端服务域
- **components/** 仅存放全局可复用组件，页面私有组件放在视图目录下的 `comps/` 子目录
- **composables/** 按业务模块分组，禁止创建大杂烩的 `useUtils`
- **store/** 每个模块独立目录，必须包含 `index.ts` 和 `types.d.ts`
- **views/** 一级目录即路由页面，复杂页面可内部再分层

---

## 3. 设计系统与主题规范

### 3.1 核心设计令牌（CSS 变量）
```css
/* src/assets/css/style.css */
:root {
  /* ===== Light 主题设计令牌 ===== */
  /* 背景色 */
  --bg-app: #f8fafc;           /* 应用全局背景（最浅灰蓝） */
  --bg-card: #ffffff;          /* 卡片/面板背景 */
  --bg-sidebar: #ffffff;       /* 侧边栏背景 */
  
  /* 边框色 */
  --border-color: #e2e8f0;     /* 统一边框色（slate-200） */
  
  /* 文本色 */
  --text-primary: #0f172a;     /* 主要文本（slate-900） */
  --text-secondary: #64748b;   /* 次要文本（slate-500） */
  --text-muted: #94a3b8;       /* 弱化文本（slate-400） */
  
  /* 圆角与间距 */
  --h-radius: 12px;            /* 全局统一圆角 */
  --h-padding: 16px;           /* 全局统一内边距 */
  
  /* 字号系统 */
  --font-xs: 0.75rem;          /* 12px */
  --font-sm: 0.8125rem;        /* 13px */
  --font-base: 0.875rem;       /* 14px - 全局基准 */
  --font-md: 0.9375rem;        /* 15px */
  --font-lg: 1rem;             /* 16px */
  --font-xl: 1.125rem;         /* 18px */
  
  /* 圆角系统 */
  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-full: 9999px;
  
  /* PrimeVue 主色调覆盖 */
  --p-primary-color: #334155;  /* slate-700 */
  --primary-color: #334155;
}

/* Dark 主题（预留） */
.dark, :root.dark, html.dark {
  --bg-app: #0b0f19;
  --bg-card: #111827;
  --bg-sidebar: #111827;
  --border-color: #1f293d;
  --text-primary: #f8fafc;
  --text-secondary: #94a3b8;
  --text-muted: #64748b;
  --p-primary-color: #94a3b8;
  --primary-color: #94a3b8;
}
```

### 3.2 设计风格原则
1. **扁平化现代风格**：小圆角（8-12px）、细边框（1px solid）、禁用过度阴影
2. **冷色调主导**：Slate 灰蓝系为主色调，避免鲜艳色彩
3. **间距统一**：使用 `--h-padding` 变量统一内边距，保持视觉节奏
4. **字重克制**：正文 400，标题 500-600，禁用 700+（除 LOGO）
5. **等宽字体场景**：端口号、PID、代码片段、环境变量必须使用 `.font-mono` 类

### 3.3 全局基础样式
```css
/* 重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  outline: none !important;
}

html {
  font-size: 16px;
  background-color: var(--bg-app);
  color: var(--text-primary);
}

body, #app {
  font-size: var(--font-base);
  line-height: 1.5;
  letter-spacing: 0.3px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background-color: var(--bg-app);
  color: var(--text-primary);
  overflow: hidden;
  user-select: none;
  width: 100%;
  height: 100%;
}

/* 简约滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 9999px;
}

::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* 工具类 */
.font-mono {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  letter-spacing: 0;
}
```

---

## 4. 组件开发规范

### 4.1 自定义组件命名约定
- **文件名**：kebab-case（`h-card.vue`, `h-page.vue`）
- **组件名**：PascalCase（`HCard`, `HPage`）
- **前缀规则**：全局复用组件统一 `h-` 前缀（代表 DevEnv Helper）

### 4.2 标准组件结构模板
```vue
<template>
  <div class="h-card" :class="{ fill }">
    <div v-if="title || $slots.title" class="header">
      <slot name="title">
        <span class="title">{{ title }}</span>
      </slot>
    </div>
    <div class="body">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
// 1. 显式声明组件名（用于 Vue DevTools）
defineOptions({
  name: 'HCard'
})

// 2. Props 接口定义
interface Props {
  title?: string
  fill?: boolean
}

// 3. 使用 withDefaults 提供默认值
const props = withDefaults(defineProps<Props>(), {
  title: '',
  fill: false
})

// 4. Emits 定义（如需要）
const emit = defineEmits<{
  click: [id: number]
}>()
</script>

<style scoped lang="scss">
.h-card {
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  overflow: hidden;
  
  &.fill {
    flex: 1;
    min-height: 0;
  }
  
  .header {
    height: 3.5rem;
    padding: 0 1.25rem;
    border-bottom: 1px dashed var(--border-color);
    display: flex;
    align-items: center;
    
    .title {
      font-size: var(--font-md);
      font-weight: 500;
      color: var(--text-primary);
    }
  }
  
  .body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }
}
</style>
```

### 4.3 核心基础组件需求

#### HCard（卡片容器）
- **用途**：统一的内容卡片容器
- **Props**：`title`, `fill`（是否填充父容器）
- **Slots**：`title`, `actions`, `default`

#### HPage（页面容器）
- **用途**：标准页面布局（顶部标题栏 + 滚动内容区）
- **Props**：`title`, `desc`, `padding`, `border`, `scroll`
- **Slots**：`header`, `title`, `desc`, `extra`, `default`
- **参考**：`zbzs/ele/src/render/src/components/h-page/h-page.vue`

---

## 5. 代码风格约束

### 5.1 Vue SFC 编写规范
```vue
<template>
  <!-- 1. 模板内 2 空格缩进 -->
  <!-- 2. 属性过多时换行对齐 -->
  <div
    class="container"
    :class="{ active: isActive }"
    @click="handleClick"
  >
    <!-- 3. v-if / v-for 禁止同时使用 -->
    <!-- 4. 使用 v-show 控制显示而非 v-if（频繁切换场景） -->
  </div>
</template>

<script setup lang="ts">
// ===== 导入顺序 =====
// 1. Vue 核心
import { ref, computed, onMounted } from 'vue'
// 2. 第三方库
import { useRoute, useRouter } from 'vue-router'
// 3. 项目内模块
import { useProjectStore } from '@/store/projects'
import { getProjects } from '@/api/projects'
// 4. 类型定义
import type { Project } from '@/types'

// ===== 组件声明 =====
defineOptions({
  name: 'ProjectList'  // 必须显式声明
})

// ===== 响应式状态 =====
const loading = ref(false)
const projects = ref<Project[]>([])

// ===== 计算属性 =====
const activeProjects = computed(() => 
  projects.value.filter(p => p.status === 'active')
)

// ===== 方法 =====
async function loadProjects() {
  loading.value = true
  try {
    const res = await getProjects()
    projects.value = res.data
  } finally {
    loading.value = false
  }
}

// ===== 生命周期 =====
onMounted(() => {
  loadProjects()
})
</script>

<style scoped lang="scss">
// 1. 使用 scoped 隔离样式
// 2. 优先使用 CSS 变量
// 3. BEM 命名或嵌套结构（不超过 3 层）
.container {
  padding: var(--h-padding);
  background: var(--bg-card);
  
  .item {
    &--active {
      color: var(--p-primary-color);
    }
  }
}
</style>
```

### 5.2 TypeScript 规范
```typescript
// ===== 接口定义 =====
// 优先使用 interface，仅在需要联合类型时用 type
export interface Project {
  id: number               // 主键统一 number（对应 Go int64）
  name: string
  path: string
  status: ProjectStatus    // 枚举类型
  createdAt: string        // ISO 8601 时间字符串
  updatedAt: string
}

// ===== 枚举 =====
export enum ProjectStatus {
  Active = 'active',
  Archived = 'archived',
  Deleted = 'deleted'
}

// ===== API 响应类型 =====
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// ===== 函数签名 =====
// 显式标注返回类型
export async function getProjects(): Promise<ApiResponse<Project[]>> {
  // ...
}
```

### 5.3 命名约定
| 类型 | 约定 | 示例 |
|------|------|------|
| 文件名 | kebab-case | `project-list.vue`, `use-projects.ts` |
| 组件名 | PascalCase | `ProjectList`, `HCard` |
| 变量/函数 | camelCase | `projectList`, `loadProjects` |
| 常量 | UPPER_SNAKE_CASE | `API_BASE_URL`, `MAX_RETRY` |
| 接口/类型 | PascalCase | `Project`, `ApiResponse` |
| 枚举 | PascalCase + 成员 PascalCase | `ProjectStatus.Active` |
| 私有函数 | 前缀 `_` | `_internalHelper()` |
| Composables | `use` 前缀 | `useProjects()`, `useConfig()` |
| Store | `use` 前缀 + `Store` 后缀 | `useProjectStore()` |

---

## 6. API 服务层规范

### 6.1 服务文件组织
```typescript
// src/api/projects.ts
import type { ApiResponse, Project } from '@/types'

// 基础 URL 由 bootstrap/http.ts 统一配置
const http = useHttp()  // 从引导模块获取 Axios 实例

/**
 * 获取项目列表
 */
export async function getProjects(): Promise<ApiResponse<Project[]>> {
  return http.get('/api/projects')
}

/**
 * 创建项目
 */
export async function createProject(data: {
  name: string
  path: string
}): Promise<ApiResponse<Project>> {
  return http.post('/api/projects', data)
}

/**
 * 删除项目（仅删除元数据）
 */
export async function deleteProject(id: number): Promise<ApiResponse<void>> {
  return http.delete(`/api/projects/${id}`)
}
```

### 6.2 统一错误处理
```typescript
// src/bootstrap/http.ts
import axios from 'axios'
import { useToast } from 'primevue/usetoast'

export function initHttp() {
  const http = axios.create({
    baseURL: import.meta.env.VITE_API_BASE || '',
    timeout: 15000
  })
  
  // 响应拦截器：统一错误处理
  http.interceptors.response.use(
    res => res.data,
    err => {
      const toast = useToast()
      toast.add({
        severity: 'error',
        summary: '请求失败',
        detail: err.response?.data?.message || err.message,
        life: 3000
      })
      return Promise.reject(err)
    }
  )
  
  return http
}
```

---

## 7. 路由与导航规范

### 7.1 路由配置
```typescript
// src/router/index.ts
import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '控制台' }
  },
  {
    path: '/projects',
    name: 'Projects',
    component: () => import('@/views/Projects.vue'),
    meta: { title: '项目管理' }
  }
  // ...
]

export const router = createRouter({
  history: createWebHashHistory(),  // 必须使用 Hash 模式（Wails 要求）
  routes
})

// 路由守卫：设置页面标题
router.afterEach((to) => {
  document.title = `${to.meta.title || 'DevEnv'} - DevEnv`
})
```

### 7.2 导航菜单数据结构
```typescript
// 导航项接口
export interface NavItem {
  label: string
  icon: string
  to: string
  badge?: string | number
}

// 示例数据
const navItems: NavItem[] = [
  { label: '控制台', icon: 'pi pi-th-large', to: '/dashboard' },
  { label: '项目管理', icon: 'pi pi-folder', to: '/projects' },
  { label: 'PHP 环境', icon: 'pi pi-code', to: '/php' },
  { label: '服务实例', icon: 'pi pi-server', to: '/services' },
  { label: '虚拟主机', icon: 'pi pi-globe', to: '/hosts' },
  { label: '任务作业', icon: 'pi pi-list-check', to: '/tasks' },
  { label: '全局设置', icon: 'pi pi-cog', to: '/settings' }
]
```

---

## 8. 状态管理规范（Pinia）

### 8.1 Store 标准结构
```typescript
// src/store/projects/index.ts
import { defineStore } from 'pinia'
import type { Project } from '@/types'

export interface ProjectsState {
  list: Project[]
  loading: boolean
  selectedId: number | null
}

export const useProjectsStore = defineStore('projects', {
  state: (): ProjectsState => ({
    list: [],
    loading: false,
    selectedId: null
  }),
  
  getters: {
    selectedProject: (state): Project | undefined => 
      state.list.find(p => p.id === state.selectedId),
    
    activeProjects: (state): Project[] =>
      state.list.filter(p => p.status === 'active')
  },
  
  actions: {
    async fetchProjects() {
      this.loading = true
      try {
        const res = await getProjects()
        this.list = res.data
      } finally {
        this.loading = false
      }
    },
    
    selectProject(id: number) {
      this.selectedId = id
    }
  },
  
  // 持久化配置（可选）
  persist: {
    key: 'devenv-projects',
    storage: localStorage,
    paths: ['selectedId']  // 仅持久化部分字段
  }
})
```

### 8.2 Store 使用规范
- **命名**：`use<Module>Store`
- **职责单一**：每个 Store 对应一个业务域
- **避免嵌套**：禁止 Store 之间互相引用（通过 Composables 协调）
- **持久化选择性**：仅持久化用户偏好，不持久化列表数据

---

## 9. Composables 组合式函数规范

### 9.1 标准 Composable 模板
```typescript
// src/composables/projects/index.ts
import { ref, computed } from 'vue'
import { useProjectsStore } from '@/store/projects'
import { useToast } from 'primevue/usetoast'
import { getProjects, createProject, deleteProject } from '@/api/projects'
import type { Project } from '@/types'

export function useProjects() {
  const store = useProjectsStore()
  const toast = useToast()
  
  // 本地状态
  const creating = ref(false)
  const deleting = ref(false)
  
  // 计算属性
  const projects = computed(() => store.list)
  const loading = computed(() => store.loading)
  
  // 方法
  async function loadProjects() {
    await store.fetchProjects()
  }
  
  async function addProject(data: { name: string; path: string }) {
    creating.value = true
    try {
      await createProject(data)
      await loadProjects()
      toast.add({
        severity: 'success',
        summary: '创建成功',
        detail: `项目 ${data.name} 已添加`,
        life: 3000
      })
    } catch (err) {
      // 错误已由 HTTP 拦截器处理
    } finally {
      creating.value = false
    }
  }
  
  async function removeProject(id: number) {
    deleting.value = true
    try {
      await deleteProject(id)
      await loadProjects()
      toast.add({
        severity: 'info',
        summary: '删除成功',
        detail: '项目记录已移除',
        life: 3000
      })
    } finally {
      deleting.value = false
    }
  }
  
  return {
    // 状态
    projects,
    loading,
    creating,
    deleting,
    // 方法
    loadProjects,
    addProject,
    removeProject
  }
}
```

### 9.2 Composables 设计原则
- **命名**：`use<Feature>` 格式
- **返回对象**：统一返回 `{ state, computed, methods }` 结构
- **职责清晰**：编排 Store + API 调用 + Toast 通知
- **避免副作用**：不在 Composable 内自动触发数据加载（除非明确需要）

---

## 10. PrimeVue 组件样式覆盖规范

### 10.1 全局覆盖原则
```css
/* src/assets/css/primevue.css */

/* ===== 扁平化 Button ===== */
.p-button .p-button-label {
  font-weight: 400 !important;
  letter-spacing: 0.5px;
}

/* ===== 扁平化 Tag 徽章 ===== */
.p-tag {
  box-shadow: none !important;
  font-weight: 500 !important;
  border: 1px solid var(--border-color) !important;
  background: var(--bg-app) !important;
  color: var(--text-secondary) !important;
}

/* ===== 精简 DataTable ===== */
.p-datatable .p-datatable-thead > tr > th {
  background: var(--bg-card) !important;
  border-bottom: 1px solid var(--border-color) !important;
  font-size: var(--font-sm) !important;
  font-weight: 500 !important;
  color: var(--text-secondary) !important;
  padding: 0.75rem 1rem !important;
}

.p-datatable .p-datatable-tbody > tr > td {
  padding: 0.875rem 1rem !important;
  font-size: var(--font-base) !important;
  border-bottom: 1px solid var(--border-color) !important;
}

/* ===== 精简 Dialog ===== */
.p-dialog .p-dialog-header {
  padding: 1.25rem 1.5rem !important;
  border-bottom: 1px solid var(--border-color) !important;
}

.p-dialog .p-dialog-content {
  padding: 1.5rem !important;
}

/* ===== 扁平化 Card（如果使用） ===== */
.p-card {
  box-shadow: none !important;
  border: 1px solid var(--border-color) !important;
  border-radius: var(--h-radius) !important;
}
```

### 10.2 组件级覆盖
```vue
<template>
  <DataTable :value="projects" class="compact-table">
    <!-- ... -->
  </DataTable>
</template>

<style scoped>
/* 组件级覆盖使用 :deep() */
.compact-table :deep(.p-datatable-tbody > tr > td) {
  padding: 0.5rem 0.75rem !important;
  font-size: var(--font-sm) !important;
}
</style>
```

---

## 11. 构建与性能优化

### 11.1 Vite 配置优化
```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  
  build: {
    target: 'esnext',
    minify: 'esbuild',
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: {
          'primevue': ['primevue'],
          'vendor': ['vue', 'vue-router', 'pinia']
        }
      }
    }
  },
  
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'primevue']
  }
})
```

### 11.2 代码分割策略
- **路由懒加载**：所有视图组件使用 `() => import()` 动态导入
- **组件懒加载**：大型第三方组件（如图表库）按需导入
- **闲时预载**：在主页面使用 `requestIdleCallback` 预载次要视图

```typescript
// views/main/index.vue
onMounted(() => {
  requestIdleCallback(() => {
    import('@/views/Projects.vue')
    import('@/views/Services.vue')
  })
})
```

---

## 12. 国际化支持（可选）

### 12.1 Vue I18n 集成（如需要）
```typescript
// src/i18n/index.ts
import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

export const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  }
})

// 全局 $t 函数（自动导入）
export const $t = i18n.global.t
```

### 12.2 语言文件结构
```typescript
// src/i18n/locales/zh-CN.ts
export default {
  dashboard: '控制台',
  projects: '项目管理',
  services: '服务实例',
  // 嵌套结构
  project: {
    create: '创建项目',
    delete: '删除项目',
    deleteConfirm: '确定要删除项目 {name} 吗？'
  }
}
```

---

## 13. 开发与调试规范

### 13.1 环境变量
```env
# .env.development
VITE_API_BASE=http://localhost:9527
VITE_APP_TITLE=DevEnv (开发模式)

# .env.production
VITE_API_BASE=
VITE_APP_TITLE=DevEnv
```

### 13.2 控制台输出规范
- **生产环境禁用 console**：Vite 配置 `drop: ['console', 'debugger']`
- **开发环境分级日志**：
  ```typescript
  console.log('[项目管理] 加载项目列表')
  console.warn('[服务管理] 端口 8080 已被占用')
  console.error('[HTTP] 请求失败:', err)
  ```

---

## 14. Git 提交规范

### 14.1 Commit Message 格式
```
<type>(<scope>): <subject>

[optional body]

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>
```

### 14.2 类型约定
- `feat`: 新功能
- `fix`: 修复 Bug
- `refactor`: 重构（不改变功能）
- `style`: 样式调整（不影响逻辑）
- `chore`: 构建配置、依赖更新
- `docs`: 文档更新

### 14.3 示例
```
feat(projects): 实现项目列表分页与搜索

- 添加 Paginator 组件集成
- 实现客户端搜索过滤
- 优化表格性能（虚拟滚动）

Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>
```

---

## 15. 迁移检查清单

从现有 `dev-env/frontend` 迁移到本规范：

- [ ] 安装 `unplugin-auto-import` 和 `unplugin-vue-components`
- [ ] 配置 Vite 插件（自动导入 + 组件解析）
- [ ] 创建 `src/assets/css/style.css` 和 `primevue.css`
- [ ] 替换 PrimeVue Aura 主题为自定义 Light 主题
- [ ] 重构目录结构（添加 `composables/`, `bootstrap/`, `components/` 等）
- [ ] 创建 `HCard` 和 `HPage` 基础组件
- [ ] 统一 API 服务层（`src/api/` 按模块拆分）
- [ ] 配置 Pinia 持久化插件
- [ ] 添加全局错误处理（HTTP 拦截器 + Toast）
- [ ] 优化 App.vue 根组件布局
- [ ] 添加路由守卫与页面标题管理
- [ ] 配置代码分割与预载策略

---

## 附录：关键参考文件

1. **样式系统**：`zbzs/ele/src/render/src/assets/css/style.css`
2. **PrimeVue 覆盖**：`zbzs/ele/src/render/src/assets/css/primevue.css`
3. **HCard 组件**：`zbzs/ele/src/render/src/components/h-card/h-card.vue`
4. **HPage 组件**：`zbzs/ele/src/render/src/components/h-page/h-page.vue`
5. **路由配置**：`zbzs/ele/src/render/src/router/index.ts`
6. **启动引导**：`zbzs/ele/src/render/src/bootstrap/index.ts`
7. **Vite 配置**：`zbzs/ele/ele.config.ts` (renderer 部分)

---

**文档版本**：v1.0.0  
**最后更新**：2026-10-14  
**维护者**：DevEnv Team
