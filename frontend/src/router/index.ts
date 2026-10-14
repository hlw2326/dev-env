import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Projects from '../views/Projects.vue'
import PHP from '../views/PHP.vue'
import Services from '../views/Services.vue'
import VirtualHosts from '../views/VirtualHosts.vue'
import Tasks from '../views/Tasks.vue'
import Settings from '../views/Settings.vue'

const routes = [
  { path: '/', name: '控制台', component: Dashboard },
  { path: '/projects', name: '项目管理', component: Projects },
  { path: '/php', name: 'PHP 环境', component: PHP },
  { path: '/services', name: '服务实例', component: Services },
  { path: '/hosts', name: '虚拟主机', component: VirtualHosts },
  { path: '/tasks', name: '任务作业', component: Tasks },
  { path: '/settings', name: '全局设置', component: Settings },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})
