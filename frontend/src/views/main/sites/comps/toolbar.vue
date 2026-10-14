<template>
	<!-- 顶部搜索、目录定位与类型筛选栏 -->
	<div class="filter-bar">
		<!-- 搜索输入与状态统计 -->
		<div class="filter-top">
			<div class="search-row flex-1">
				<div class="search-box">
					<i class="pi pi-search search-icon"></i>
					<p-input-text
						v-model="keyword"
						placeholder="搜索网站名称、绑定域名或根目录路径..."
						class="search-input"
					/>
				</div>
				<p-button
					label="选择目录"
					icon="pi pi-folder-open"
					severity="secondary"
					outlined
					@click="pickSearchDir"
				/>
			</div>

			<div class="sum-badges font-mono">
				<span class="sum-badge">总计: {{ sites.length }}</span>
				<span class="sum-badge running">运行中: {{ runningCount }}</span>
				<span class="sum-badge stopped">已停止: {{ sites.length - runningCount }}</span>
			</div>
		</div>

		<!-- 网站类型筛选页签 -->
		<div class="filter-tabs">
			<button
				v-for="tab in kindTabs"
				:key="tab.key"
				type="button"
				class="type-tab"
				:class="{ active: curKind === tab.key }"
				@click="curKind = tab.key"
			>
				<i :class="tab.icon"></i>
				<span>{{ tab.label }}</span>
				<span class="tab-count font-mono">{{ getKindCount(tab.key) }}</span>
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
	import {
		kindTabs,
		getKindCount,
		useSites
	} from '../use-sites'

	// 解构搜索与筛选状态
	const {
		sites,
		keyword,
		curKind,
		runningCount,
		pickSearchDir
	} = useSites()
</script>

<style scoped lang="scss">
	/* 筛选工具栏容器 */
	.filter-bar {
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--h-radius);
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	/* 顶部搜索与统计行 */
	.filter-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
	}

	/* 搜索输入与按钮行 */
	.search-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 搜索框容器 */
	.search-box {
		position: relative;
		display: flex;
		align-items: center;
		width: 340px;
	}

	/* 搜索图标 */
	.search-icon {
		position: absolute;
		left: 10px;
		color: var(--text-muted);
		font-size: var(--font-sm);
		pointer-events: none;
	}

	/* 搜索输入框 */
	.search-input {
		width: 100%;
		padding-left: 32px;
	}

	/* 自适应宽度 */
	.flex-1 {
		flex: 1;
	}

	/* 统计徽章组 */
	.sum-badges {
		display: flex;
		gap: 8px;
	}

	/* 单个统计徽章 */
	.sum-badge {
		font-size: var(--font-xs);
		padding: 4px 10px;
		border-radius: var(--radius-sm);
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		color: var(--text-secondary);

		&.running {
			background: var(--ok-bg);
			color: var(--ok);
			border-color: var(--ok-border);
		}
	}

	/* 类型切换按钮行 */
	.filter-tabs {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}

	/* 类型切换按钮 */
	.type-tab {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border-color);
		background: var(--bg-app);
		color: var(--text-secondary);
		font-size: var(--font-xs);
		cursor: pointer;

		&.active {
			background: var(--p-primary-color);
			color: var(--on-primary);
			border-color: var(--p-primary-color);
		}
	}

	/* 页签数量标签 */
	.tab-count {
		opacity: 0.8;
	}
</style>
