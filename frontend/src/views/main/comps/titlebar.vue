<template>
	<header class="titlebar">
		<!-- 左侧品牌与当前页面名 -->
		<div class="left">
			<div class="logo">
				<i class="pi pi-box"></i>
			</div>
			<span class="brand">DevEnv</span>
			<span class="sep">/</span>
			<span class="page font-mono">{{ routeName }}</span>
		</div>

		<!-- 中间拖拽区与应用标题 -->
		<div class="center" @dblclick="toggleMaxWin">
			<span class="title">{{ $t('本地集成开发环境管理平台') }}</span>
		</div>

		<!-- 右侧环境标识与窗口控制 -->
		<div class="right">
			<span class="env font-mono">{{ $t('Win64 · Go & Wails') }}</span>
			<win-btns />
		</div>
	</header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { $t } from '@/i18n'
import { toggleMaxWin } from '@/api'
import WinBtns from './win-btns.vue'

// 当前路由对象
const route = useRoute()

// 当前页面名称
const routeName = computed(() => {
	if (route.meta.title) {
		return String(route.meta.title)
	}
	return String(route.name)
})
</script>

<style scoped lang="scss">
/* 顶部自定义标题栏 */
.titlebar {
	height: 2.375rem;
	background: var(--bg-card);
	border-bottom: 1px solid var(--border-color);
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-left: 0.625rem;
	user-select: none;
	--wails-draggable: drag;
	flex-shrink: 0;
	z-index: 1000;
}

/* 左侧品牌区 */
.left {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	--wails-draggable: drag;

	.logo {
		width: 1.375rem;
		height: 1.375rem;
		border-radius: var(--radius-sm);
		background: var(--p-primary-color);
		color: var(--on-primary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--font-xs);
	}

	.brand {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	.sep {
		color: var(--border-hover);
		font-size: var(--font-sm);
	}

	.page {
		font-size: var(--font-sm);
		color: var(--text-secondary);
	}
}

/* 中间拖拽标题区 */
.center {
	flex: 1;
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	--wails-draggable: drag;

	.title {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
}

/* 右侧控制区 */
.right {
	display: flex;
	align-items: center;
	height: 100%;
	gap: 0.75rem;

	.env {
		font-size: var(--font-xs);
		color: var(--text-secondary);
		background: var(--bg-app);
		padding: 0.125rem 0.5rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border-color);
		--wails-draggable: no-drag;
	}
}
</style>
