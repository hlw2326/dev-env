<template>
	<aside class="sidebar">
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
const navItems = [
	{ label: '环境概览', icon: 'pi pi-objects-column', to: '/' },
	{ label: '网站管理', icon: 'pi pi-globe', to: '/sites' },
	{ label: '软件商店', icon: 'pi pi-shopping-bag', to: '/software' },
	{ label: '全局设置', icon: 'pi pi-cog', to: '/settings' },
]
</script>

<style scoped lang="scss">
.sidebar {
	width: 11.875rem;
	background: var(--bg-sidebar);
	border-right: 1px solid var(--border-color);
	display: flex;
	flex-direction: column;
}

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
			font-size: var(--icon-md);
		}
	}
}
</style>
