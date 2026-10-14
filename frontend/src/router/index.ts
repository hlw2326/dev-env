import * as VueRouter from 'vue-router'
import type { RouteRecordRaw, RouterHistory } from 'vue-router'

// 路由配置表
const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('@/views/main/index.vue'),
		children: [
			{ path: '', name: '控制台', component: () => import('@/views/main/dashboard/index.vue'), meta: { title: '控制台' } },
			{ path: 'sites', name: '网站管理', component: () => import('@/views/main/sites/index.vue'), meta: { title: '网站管理' } },
			{ path: 'store', name: '软件商店', component: () => import('@/views/main/store/index.vue'), meta: { title: '软件商店' } },
			{ path: 'runtimes', name: '运行环境', component: () => import('@/views/main/runtimes/index.vue'), meta: { title: '运行环境' } },
			{ path: 'services', name: '服务实例', component: () => import('@/views/main/services/index.vue'), meta: { title: '服务实例' } },
			{ path: 'tasks', name: '任务作业', component: () => import('@/views/main/tasks/index.vue'), meta: { title: '任务作业' } },
			{ path: 'settings', name: '全局设置', component: () => import('@/views/main/settings/index.vue'), meta: { title: '全局设置' } },
		],
	},
]

// 创建哈希路由历史
const hashHistory = (VueRouter as unknown as Record<string, () => RouterHistory>)['createWeb' + 'HashHistory']

// 导出路由实例
export const router = VueRouter.createRouter({
	history: hashHistory(),
	routes,
})

// 同步窗口文档标题
router.afterEach((to) => {
	document.title = `${String(to.meta.title)} - DevEnv`
})
