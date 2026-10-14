<template>
	<aside class="sidebar">
		<!-- 导航菜单列表 -->
		<nav class="nav">
			<router-link
				v-for="item in navItems"
				:key="item.to"
				:to="item.to"
				class="item"
				exact-active-class="active"
			>
				<i :class="[item.icon, 'icon']"></i>
				<span>{{ item.label }}</span>
			</router-link>
		</nav>
	</aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { $t } from '@/i18n'

// 侧边导航菜单项
const navItems = computed(() => [
	{ label: $t('控制台'), icon: 'pi pi-th-large', to: '/' },
	{ label: $t('网站管理'), icon: 'pi pi-globe', to: '/sites' },
	{ label: $t('软件商店'), icon: 'pi pi-shopping-bag', to: '/store' },
	{ label: $t('运行环境'), icon: 'pi pi-box', to: '/runtimes' },
	{ label: $t('服务实例'), icon: 'pi pi-server', to: '/services' },
	{ label: $t('任务作业'), icon: 'pi pi-list-check', to: '/tasks' },
	{ label: $t('全局设置'), icon: 'pi pi-cog', to: '/settings' },
])
</script>

<style scoped lang="scss">
/* 左侧导航栏容器 */
.sidebar {
	width: 11.875rem;
	background: var(--bg-sidebar);
	border-right: 1px solid var(--border-color);
	display: flex;
	flex-direction: column;
}

/* 导航链接列表 */
.nav {
	display: flex;
	flex-direction: column;
	flex: 1;
	padding: 0.375rem 0;

	.item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		height: 2.75rem;
		padding: 0 1.125rem;
		color: var(--text-secondary);
		text-decoration: none;
		font-size: var(--font-sm);
		border-right: 3px solid transparent;

		&:hover {
			background: var(--bg-app);
			color: var(--text-primary);
		}

		&.active {
			background: var(--bg-app);
			color: var(--text-primary);
			border-right-color: var(--p-primary-color);
		}

		.icon {
			font-size: var(--font-md);
		}
	}
}
</style>
