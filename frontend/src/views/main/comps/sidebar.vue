<template>
	<!-- 左侧导航栏 -->
	<aside class="sidebar">
		<!-- 导航菜单列表 -->
		<nav class="menu">
			<router-link
				v-for="item in menuItems"
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
// 侧边栏导航项列表
const menuItems = [
	{ label: '环境概览', icon: 'pi pi-objects-column', to: '/' },
	{ label: '项目管理', icon: 'pi pi-folder-open', to: '/projects' },
	{ label: '服务管理', icon: 'pi pi-server', to: '/services' },
	{ label: '软件商店', icon: 'pi pi-shopping-bag', to: '/software' },
	{ label: '全局设置', icon: 'pi pi-cog', to: '/settings' },
]
</script>

<style scoped lang="scss">
/* 侧边栏容器 */
.sidebar {
	width: 11.875rem;
	background: var(--bg-sidebar);
	border-right: 1px solid var(--border-color);
	display: flex;
	flex-direction: column;
}

/* 菜单列表区域 */
.menu {
	display: flex;
	flex-direction: column;
	flex: 1;
	padding: 0.375rem 0;

	/* 单个导航项 */
	.item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		height: 2.75rem;
		padding: 0 1.125rem;
		color: var(--text-secondary);
		text-decoration: none;
		font-size: var(--font-sm);
		border-right: 0.1875rem solid transparent;

		/* 悬停态 */
		&:hover {
			background: var(--bg-app);
			color: var(--text-primary);
		}

		/* 激活态 */
		&.active {
			background: var(--bg-app);
			color: var(--text-primary);
			border-right-color: var(--p-primary-color);
		}

		/* 导航图标 */
		.icon {
			font-size: var(--icon-md);
		}
	}
}
</style>
