import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', name: '控制台', component: () => import('@/views/dashboard.vue'), meta: { title: '控制台' } },
  { path: '/projects', name: '项目管理', component: () => import('@/views/projects.vue'), meta: { title: '项目管理' } },
  { path: '/store', name: '软件商店', component: () => import('@/views/store.vue'), meta: { title: '软件商店' } },
  { path: '/runtimes', name: '运行环境', component: () => import('@/views/runtimes.vue'), meta: { title: '运行环境' } },
  { path: '/php', redirect: '/runtimes' },
  { path: '/services', name: '服务实例', component: () => import('@/views/services.vue'), meta: { title: '服务实例' } },
  { path: '/hosts', name: '虚拟主机', component: () => import('@/views/virtual-hosts.vue'), meta: { title: '虚拟主机' } },
  { path: '/tasks', name: '任务作业', component: () => import('@/views/tasks.vue'), meta: { title: '任务作业' } },
  { path: '/settings', name: '全局设置', component: () => import('@/views/settings.vue'), meta: { title: '全局设置' } },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.afterEach((to) => {
  document.title = `${String(to.meta.title || 'DevEnv')} - DevEnv`
})
