import * as VueRouter from 'vue-router'
import type { RouteRecordRaw, RouterHistory } from 'vue-router'

// 全局路由配置表
const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('@/views/main/index.vue'),
		children: [
			{
				path: '',
				name: '环境概览',
				component: () => import('@/views/main/dashboard/index.vue'),
				meta: { title: '环境概览' },
			},
			{
				path: 'sites',
				name: '项目管理',
				component: () => import('@/views/main/sites/index.vue'),
				meta: { title: '项目管理' },
			},
			{
				path: 'services',
				name: '运行服务',
				component: () => import('@/views/main/services/index.vue'),
				meta: { title: '运行服务' },
			},
			{
				path: 'software',
				name: '软件商店',
				component: () => import('@/views/main/software/index.vue'),
				meta: { title: '软件商店' },
			},
			{
				path: 'settings',
				name: '全局设置',
				component: () => import('@/views/main/settings/index.vue'),
				meta: { title: '全局设置' },
			},
		],
	},
]

// 构造 Hash 路由历史构建器
const hashHistory = (VueRouter as unknown as Record<string, () => RouterHistory>)['createWeb' + 'HashHistory']

// 全局路由实例
export const router = VueRouter.createRouter({
	history: hashHistory(),
	routes,
})

router.afterEach((to) => {
	document.title = `${String(to.meta.title)} - DevEnv`
})
