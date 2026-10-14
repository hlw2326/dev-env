import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', name: '控制台', component: () => import('@/views/Dashboard.vue'), meta: { title: '控制台' } },
  { path: '/projects', name: '项目管理', component: () => import('@/views/Projects.vue'), meta: { title: '项目管理' } },
  { path: '/php', name: 'PHP 环境', component: () => import('@/views/PHP.vue'), meta: { title: 'PHP 环境' } },
  { path: '/services', name: '服务实例', component: () => import('@/views/Services.vue'), meta: { title: '服务实例' } },
  { path: '/hosts', name: '虚拟主机', component: () => import('@/views/VirtualHosts.vue'), meta: { title: '虚拟主机' } },
  { path: '/tasks', name: '任务作业', component: () => import('@/views/Tasks.vue'), meta: { title: '任务作业' } },
  { path: '/settings', name: '全局设置', component: () => import('@/views/Settings.vue'), meta: { title: '全局设置' } },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

router.afterEach((to) => {
  document.title = `${String(to.meta.title || 'DevEnv')} - DevEnv`
})
