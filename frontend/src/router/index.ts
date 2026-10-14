import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Projects from '../views/Projects.vue'
import PHP from '../views/PHP.vue'
import Services from '../views/Services.vue'
import VirtualHosts from '../views/VirtualHosts.vue'
import Tasks from '../views/Tasks.vue'
import Settings from '../views/Settings.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/projects', name: 'Projects', component: Projects },
  { path: '/php', name: 'PHP', component: PHP },
  { path: '/services', name: 'Services', component: Services },
  { path: '/hosts', name: 'VirtualHosts', component: VirtualHosts },
  { path: '/tasks', name: 'Tasks', component: Tasks },
  { path: '/settings', name: 'Settings', component: Settings },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})
