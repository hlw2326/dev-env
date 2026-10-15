<template>
	<!-- 项目筛选与搜索工具栏 -->
	<div class="toolbar surface-card">
		<!-- 项目类型筛选按钮组 -->
		<div class="filter-tabs">
			<button
				v-for="tab in filterOptions"
				:key="tab.value"
				type="button"
				class="filter-btn"
				:class="{ active: kindFilter === tab.value }"
				@click="kindFilter = tab.value"
			>
				{{ tab.label }}
			</button>
		</div>

		<!-- 项目名称与域名搜索框 -->
		<div class="search-box">
			<i class="pi pi-search search-icon"></i>
			<p-input-text
				v-model="search"
				placeholder="搜索项目名称或域名..."
				class="search-input"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useProjects } from '../use-projects'
import type { ProjectTemplate } from '@/types'

const props = defineProps<{
	templates: ProjectTemplate[]
}>()

// 筛选与搜索状态
const { search, kindFilter } = useProjects()

// 项目类型筛选项来自后端模板注册表
const filterOptions = computed(() => [
	{ value: 'all', label: '全部' },
	...props.templates.map((item) => ({ value: item.kind, label: item.label })),
])
</script>

<style scoped lang="scss">
/* 顶部工具栏容器 */
.toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	gap: 1rem;
}

/* 筛选页签组 */
.filter-tabs {
	display: flex;
	align-items: center;
	gap: 0.375rem;
}

/* 单个筛选按钮 */
.filter-btn {
	border: 1px solid transparent;
	background: transparent;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	padding: 0.35rem 0.75rem;
	border-radius: var(--h-radius);
	cursor: pointer;
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		background: var(--bg-app);
		color: var(--text-primary);
	}

	/* 选中态 */
	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
	}
}

/* 搜索框容器 */
.search-box {
	position: relative;
	display: flex;
	align-items: center;
	width: 16rem;

	/* 搜索前缀图标 */
	.search-icon {
		position: absolute;
		left: 0.75rem;
		font-size: var(--icon-sm);
		color: var(--text-muted);
	}

	/* 搜索文本输入框 */
	.search-input {
		width: 100%;
		padding-left: 2.15rem !important;
	}
}
</style>
